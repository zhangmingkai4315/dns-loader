package core

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"sync/atomic"
	"time"
)

type myLoaderGenerator struct {
	caller        Caller
	timeout       time.Duration
	qps           uint32
	status        uint32
	duration      time.Duration
	pool          LoaderTicketPool
	ctx           context.Context
	cancelFunc    context.CancelFunc
	callCount     int64
	concurrency   uint32
	resultChannel chan *CallResult
}

func (mlg *myLoaderGenerator) init() {
	var buf bytes.Buffer
	buf.WriteString("Initial dns loader...")
	var total = int64(mlg.timeout)/int64(1e9/mlg.qps) + 1
	if total > math.MaxInt32 {
		total = math.MaxInt32
	}
	mlg.concurrency = uint32(total)
	pool := NewLoaderPool(mlg.concurrency)
	mlg.pool = pool

	buf.WriteString(fmt.Sprintf("Initial Process Done QPS[%d]/concurrency[%d]", mlg.qps, mlg.concurrency))
}

func (mlg *myLoaderGenerator) Start() bool {
	log.Println("Starting Loader...")
	mlg.ctx, mlg.cancelFunc = context.WithTimeout(context.Background(), mlg.duration)
	mlg.callCount = 0
	currentStatus := mlg.Status()
	if currentStatus != STATUS_STARTING && currentStatus != STATUS_STOPPED {
		return false
	}
	atomic.StoreUint32(&mlg.status, STATUS_STARTING)

	var throttle <-chan time.Time
	if mlg.qps > 0 {
		interval := time.Duration(1e9 / mlg.qps)
		log.Printf("Setting Throttle %v", interval)
		throttle = time.Tick(interval)
	}

	atomic.StoreUint32(&mlg.status, STATUS_STARTED)
	log.Println("Setting Done For Loader")

	go func() {
		log.Println("Create New Goroutine TO Generating Load")
		mlg.generatorLoad(throttle)
		log.Printf("Stopped [%d]", mlg.CallCount())
	}()

	return true
}

func (mlg *myLoaderGenerator) prepareStop(err error) {
	log.Printf("Prepare to Stop Load Test [%s]\n", err)
	atomic.StoreUint32(&mlg.status, STATUS_STOPPING)
	log.Println("Stop Channel...")
	close(mlg.resultChannel)
	atomic.StoreUint32(&mlg.status, STATUS_STOPPED)
	log.Println("Stop Load Test Done!")
}

func (mlg *myLoaderGenerator) showIgnore(result *CallResult, err string) {
	resultMsg := fmt.Sprintf(
		"ID=%d, Code=%d, Msg=%s, Elapse=%v",
		result.ID, result.Code, result.Msg, result.Elapse)
	log.Printf("Ignored result: %s. (Error: %s)\n", resultMsg, err)
}

func (mlg *myLoaderGenerator) sendNewRequest() {
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
				result := &CallResult{
					ID:   -1,
					Code: RET_CALL_ERROR,
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
			if !atomic.CompareAndSwapUint32(&callStatus, CALL_NOT_FINISH, CALL_TIMEOUT) {
				return
			}
			result := &CallResult{
				ID:     rawRequest.ID,
				Req:    rawRequest,
				Code:   RET_TIMEOUT,
				Elapse: mlg.timeout,
			}
			mlg.collectResult(result)
		})
		rawResponse := mlg.CallFunc(&rawRequest)
		if !atomic.CompareAndSwapUint32(&callStatus, CALL_NOT_FINISH, CALL_SUCCESS_DONE) {
			return
		}
		timer.Stop()
		var result *CallResult
		if rawResponse.Err != nil {
			result = &CallResult{
				ID:     rawResponse.ID,
				Req:    rawRequest,
				Code:   RET_CALL_ERROR,
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

func (mlg *myLoaderGenerator) CallFunc(rawReq *RawRequest) *RawResponse {
	atomic.AddInt64(&mlg.callCount, 1)
	if rawReq == nil {
		return &RawResponse{ID: -1, Err: errors.New("Invalid Request Raw Data")}
	}
	start := time.Now().UnixNano()
	resp, err := mlg.caller.Call(rawReq.Req, mlg.timeout)
	end := time.Now().UnixNano()

	elapsedTime := time.Duration(end - start)
	var rawResponse RawResponse
	if err != nil {
		rawResponse = RawResponse{
			ID:     rawReq.ID,
			Err:    errors.New(err.Error()),
			Elapse: elapsedTime,
		}
	} else {
		rawResponse = RawResponse{
			ID:     rawReq.ID,
			Resp:   resp,
			Elapse: elapsedTime,
		}
	}
	return &rawResponse
}

func (mlg *myLoaderGenerator) collectResult(result *CallResult) bool {
	if atomic.LoadUint32(&mlg.status) != STATUS_STARTED {
		mlg.showIgnore(result, "Stopped Loader Test")
		return false
	}
	select {
	case mlg.resultChannel <- result:
		return true
	default:
		mlg.showIgnore(result, "Full Result Channel")
		return false
	}
}

func (mlg *myLoaderGenerator) generatorLoad(throttle <-chan time.Time) {
	for {
		select {
		case <-mlg.ctx.Done():
			mlg.prepareStop(mlg.ctx.Err())
			return
		default:
		}
		mlg.sendNewRequest()
		select {
		// Only get value from throttle , do nothing till next for loop
		case <-throttle:
		case <-mlg.ctx.Done():
			mlg.prepareStop(mlg.ctx.Err())
			return
		}

	}
}

func (mlg *myLoaderGenerator) Stop() bool {
	if !atomic.CompareAndSwapUint32(
		&mlg.status, STATUS_STARTED, STATUS_STOPPING) {
		return false
	}
	mlg.cancelFunc()
	for {
		if atomic.LoadUint32(&mlg.status) == STATUS_STOPPED {
			break
		}
		time.Sleep(time.Microsecond)
	}
	return true
}

func (mlg *myLoaderGenerator) Status() uint32 {
	return atomic.LoadUint32(&mlg.status)
}
func (mlg *myLoaderGenerator) CallCount() int64 {
	return atomic.LoadInt64(&mlg.callCount)
}

// NewLoaderGenerator will return a new instance of generator
// using param from GeneratorParam
func NewLoaderGenerator(param GeneratorParam) (Generator, error) {
	log.Println("New Load Generator")
	if err := param.ValidCheck(); err != nil {
		return nil, err
	}
	mlg := &myLoaderGenerator{
		caller:        param.Caller,
		timeout:       param.Timeout,
		qps:           param.QPS,
		duration:      param.Duration,
		status:        STATUS_ORIGINAL,
		resultChannel: param.ResultChannel,
	}
	mlg.init()
	return mlg, nil
}
