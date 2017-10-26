package dnsloader

import (
	"context"
	"errors"
	"fmt"
	"github.com/briandowns/spinner"
	"github.com/zhangmingkai4315/dns-loader/core"
	"log"
	"math"
	"sync/atomic"
	"time"
)

type myDNSLoaderGenerator struct {
	caller        core.Caller
	timeout       time.Duration
	qps           uint32
	status        uint32
	duration      time.Duration
	pool          core.LoaderTicketPool
	ctx           context.Context
	cancelFunc    context.CancelFunc
	callCount     int64
	concurrency   uint32
	resultChannel chan *core.CallResult
}

func (mlg *myDNSLoaderGenerator) init() {
	log.Println("Initial common loader...")
	var total = int64(mlg.timeout)/int64(1e9/mlg.qps) + 1
	if total > math.MaxInt32 {
		total = math.MaxInt32
	}
	mlg.concurrency = uint32(total)
	pool := core.NewLoaderPool(mlg.concurrency * 10)
	mlg.pool = pool
	log.Printf("Initial Process Done QPS[%d]/concurrency[%d]", mlg.qps, mlg.concurrency)
}

func (mlg *myDNSLoaderGenerator) Start() bool {
	log.Println("Starting Loader...")
	mlg.ctx, mlg.cancelFunc = context.WithTimeout(context.Background(), mlg.duration)
	mlg.callCount = 0
	currentStatus := mlg.Status()
	if currentStatus != core.STATUS_STARTING && currentStatus != core.STATUS_STOPPED {
		return false
	}
	atomic.StoreUint32(&mlg.status, core.STATUS_STARTING)

	var throttle <-chan time.Time
	if mlg.qps > 0 {
		interval := time.Duration(1e9 / mlg.qps)
		log.Printf("Setting Throttle %v", interval)
		throttle = time.Tick(interval)
	}

	atomic.StoreUint32(&mlg.status, core.STATUS_STARTED)
	log.Println("Setting Done For Loader")

	go func() {
		log.Println("Create New Goroutine TO Generating Load")
		spinnerChannel := make(chan struct{})
		go func() {
			s := spinner.New(spinner.CharSets[35], 250*time.Millisecond)
			s.Start()
			for {
				select {
				case <-spinnerChannel:
					s.Stop()
					break
				default:
					continue
				}
			}
		}()
		// for{
		// 	// mlg.caller.BuildReq()
		// 	rawRequest := mlg.caller.BuildReq()
		// 	mlg.CallFunc(&rawRequest)
		// }
		mlg.generatorLoad(throttle, spinnerChannel)
		log.Printf("Stopped [%d]", mlg.CallCount())
	}()

	return true
}

func (mlg *myDNSLoaderGenerator) prepareStop(err error) {
	log.Printf("Prepare to Stop Load Test [%s]\n", err)
	atomic.StoreUint32(&mlg.status, core.STATUS_STOPPING)
	log.Println("Stop Channel...")
	close(mlg.resultChannel)
	atomic.StoreUint32(&mlg.status, core.STATUS_STOPPED)
	log.Println("Stop Load Test Done!")
}

func (mlg *myDNSLoaderGenerator) showIgnore(result *core.CallResult, err string) {
	resultMsg := fmt.Sprintf(
		"ID=%d, Code=%d, Msg=%s, Elapse=%v",
		result.ID, result.Code, result.Msg, result.Elapse)
	log.Printf("Ignored result: %s. (Error: %s)\n", resultMsg, err)
}

func (mlg *myDNSLoaderGenerator) sendNewRequest() {
	// Get resource from pool
	mlg.pool.Get()
	go func() {
		defer func() {
			if p := recover(); p != nil {
				err, ok := interface{}(p).(error)
				var errMessage string
				if ok {
					errMessage = fmt.Sprintf("Call Panic[%s]", err)
				} else {
					errMessage = fmt.Sprintf("Call Panic[%s]", p)
				}
				log.Println(errMessage)
				result := &core.CallResult{
					ID:   -1,
					Code: core.RET_CALL_ERROR,
					Msg:  errMessage,
				}
				mlg.collectResult(result)
			}
			// Finally will return to pool
			mlg.pool.Return()
		}()

		rawRequest := mlg.caller.BuildReq()
		var callStatus uint32
		timer := time.AfterFunc(mlg.timeout, func() {
			if !atomic.CompareAndSwapUint32(&callStatus, core.CALL_NOT_FINISH, core.CALL_TIMEOUT) {
				return
			}
			result := &core.CallResult{
				ID:     rawRequest.ID,
				Req:    rawRequest,
				Code:   core.RET_TIMEOUT,
				Elapse: mlg.timeout,
			}
			mlg.collectResult(result)
		})
		rawResponse := mlg.CallFunc(&rawRequest)
		if !atomic.CompareAndSwapUint32(&callStatus, core.CALL_NOT_FINISH, core.CALL_SUCCESS_DONE) {
			return
		}
		timer.Stop()
		var result *core.CallResult
		if rawResponse.Err != nil {
			result = &core.CallResult{
				ID:     rawResponse.ID,
				Req:    rawRequest,
				Code:   core.RET_CALL_ERROR,
				Msg:    rawResponse.Err.Error(),
				Elapse: rawResponse.Elapse,
			}
		} else {
			result = mlg.caller.CheckResp(rawRequest, *rawResponse)
			result.Elapse = rawResponse.Elapse
		}
		mlg.collectResult(result)
	}()
}

func (mlg *myDNSLoaderGenerator) CallFunc(rawReq *core.RawRequest) *core.RawResponse {
	atomic.AddInt64(&mlg.callCount, 1)
	if rawReq == nil {
		return &core.RawResponse{ID: -1, Err: errors.New("Invalid Request Raw Data")}
	}
	start := time.Now().UnixNano()
	resp, err := mlg.caller.Call(rawReq.Req, mlg.timeout)
	end := time.Now().UnixNano()

	elapsedTime := time.Duration(end - start)
	var rawResponse core.RawResponse
	if err != nil {
		rawResponse = core.RawResponse{
			ID:     rawReq.ID,
			Err:    errors.New(err.Error()),
			Elapse: elapsedTime,
		}
	} else {
		rawResponse = core.RawResponse{
			ID:     rawReq.ID,
			Resp:   resp,
			Elapse: elapsedTime,
		}
	}
	return &rawResponse
}

func (mlg *myDNSLoaderGenerator) collectResult(result *core.CallResult) bool {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Maybe channel already shutdown for receive data", r.(error).Error())
		}
	}()
	if atomic.LoadUint32(&mlg.status) != core.STATUS_STARTED {
		mlg.showIgnore(result, "Stopped Loader Test")
		return false
	}
	select {
	case mlg.resultChannel <- result:
		return true
	default:
		// mlg.showIgnore(result, "Full Result Channel")
		return false
	}
}

func (mlg *myDNSLoaderGenerator) generatorLoad(throttle <-chan time.Time, spinnerChannel chan<- struct{}) {
	for {
		select {
		case <-mlg.ctx.Done():
			spinnerChannel <- struct{}{}
			mlg.prepareStop(mlg.ctx.Err())
			return
		default:
		}
		// rawRequest := mlg.caller.BuildReq()
		// mlg.CallFunc(&rawRequest)
		mlg.sendNewRequest()
		select {
		// Only get value from throttle , do nothing till next for loop
		case <-throttle:
		case <-mlg.ctx.Done():
			spinnerChannel <- struct{}{}
			mlg.prepareStop(mlg.ctx.Err())
			return
		}

	}
}

func (mlg *myDNSLoaderGenerator) Stop() bool {
	if !atomic.CompareAndSwapUint32(
		&mlg.status, core.STATUS_STARTED, core.STATUS_STOPPING) {
		return false
	}
	mlg.cancelFunc()
	for {
		if atomic.LoadUint32(&mlg.status) == core.STATUS_STOPPED {
			break
		}
		time.Sleep(time.Microsecond)
	}
	return true
}

func (mlg *myDNSLoaderGenerator) Status() uint32 {
	return atomic.LoadUint32(&mlg.status)
}
func (mlg *myDNSLoaderGenerator) CallCount() int64 {
	return atomic.LoadInt64(&mlg.callCount)
}

// NewDNSLoaderGenerator will return a new instance of generator
// using param from GeneratorParam
func NewDNSLoaderGenerator(param core.GeneratorParam) (core.Generator, error) {
	log.Println("New Load Generator")
	if err := param.ValidCheck(); err != nil {
		return nil, err
	}
	mlg := &myDNSLoaderGenerator{
		caller:        param.Caller,
		timeout:       param.Timeout,
		qps:           param.QPS,
		duration:      param.Duration,
		status:        core.STATUS_ORIGINAL,
		resultChannel: param.ResultChannel,
	}
	mlg.init()
	return mlg, nil
}
