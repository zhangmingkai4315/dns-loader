package dnsloader

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/zhangmingkai4315/dns-loader/dns"

	"github.com/briandowns/spinner"
	log "github.com/sirupsen/logrus"
	"go.uber.org/ratelimit"
)

//GloablGenerator define  global control object
var GloablGenerator Generator
var globalStatus uint32

// GetGlobalStatus 返回当前组件的运行状态
func GetGlobalStatus() uint32 {
	return globalStatus
}

type myDNSLoaderGenerator struct {
	caller     Caller
	timeout    time.Duration
	qps        uint32
	status     uint32
	duration   time.Duration
	ctx        context.Context
	cancelFunc context.CancelFunc
	callCount  uint64
	workers    int
	startTime  time.Time
	result     map[uint8]uint64
}

func (mlg *myDNSLoaderGenerator) Start() bool {
	log.Info("starting dns loader generator")
	mlg.ctx, mlg.cancelFunc = context.WithTimeout(context.Background(), mlg.duration)
	mlg.callCount = 0
	currentStatus := mlg.Status()
	if currentStatus != STATUS_STOPPED {
		return false
	}
	atomic.StoreUint32(&mlg.status, STATUS_STARTING)
	globalStatus = STATUS_STARTING
	if mlg.qps > 0 {
		interval := time.Duration(1e9 / mlg.qps)
		log.Infof("setting throttle %v", interval)
	}
	atomic.StoreUint32(&mlg.status, STATUS_RUNNING)
	globalStatus = STATUS_RUNNING
	s := spinner.New(spinner.CharSets[36], 100*time.Millisecond)
	log.Infoln("create new thread to receive dns data from server")
	go func() {
		// recive data from connections
		b := make([]byte, 4)
		dnsclient := mlg.caller.(*DNSClient)
		for {
			n, err := dnsclient.Conn.Read(b)
			if err == nil && n > 0 {
				code := b[3] & 0x0f
				mlg.result[code] = mlg.result[code] + 1
			}
		}
	}()

	var limiter ratelimit.Limiter
	if mlg.qps > 0 {
		limiter = ratelimit.New(int(mlg.qps))
	}
	mlg.startTime = time.Now()
	mlg.generatorLoad(limiter, s)
	return true
}

func (mlg *myDNSLoaderGenerator) prepareStop(err error) {
	log.Printf("prepare to stop load test [%s]", err)
	atomic.StoreUint32(&mlg.status, STATUS_STOPPING)
	globalStatus = STATUS_STOPPING
	log.Infoln("doing calculation work")
	runningTime := time.Since(mlg.startTime)
	managerCounter := mlg.CallCount()
	log.WithFields(log.Fields{"result": true}).Infof("total packets sum:%d", managerCounter)
	log.WithFields(log.Fields{"result": true}).Infof("runing time %v", runningTime)
	var counter uint64
	for k, v := range mlg.result {
		counter = v + counter
		log.WithFields(log.Fields{"result": true}).Infof("status %s:%d [%.2f]", dns.DNSRcodeReverse[k], v, float64(v*100)/float64(mlg.CallCount()))
	}
	log.Println(mlg.CallCount(), counter)
	var unknown uint64
	if managerCounter > counter {
		unknown = managerCounter - counter
	}
	log.WithFields(log.Fields{"result": true}).Infof("status unknown:%d [%.2f]", unknown, float64(unknown*100)/float64(mlg.CallCount()))
	atomic.StoreUint32(&mlg.status, STATUS_STOPPED)
	globalStatus = STATUS_STOPPED
	log.Info("stop success!")
}

func (mlg *myDNSLoaderGenerator) sendNewRequest() {
	defer func() {
		if p := recover(); p != nil {
			err, _ := interface{}(p).(error)
			log.Println(err)
		}
	}()
	rawRequest := mlg.caller.BuildReq()
	mlg.caller.Call(rawRequest)
}

func (mlg *myDNSLoaderGenerator) generatorLoad(limiter ratelimit.Limiter, spinnerInstance *spinner.Spinner) {
	spinnerInstance.Start()
	if mlg.qps > 0 {
		for {
			select {
			case <-mlg.ctx.Done():
				spinnerInstance.Stop()
				mlg.prepareStop(mlg.ctx.Err())
				return
			default:
			}
			limiter.Take()
			rawRequest := mlg.caller.BuildReq()
			mlg.caller.Call(rawRequest)
			atomic.AddUint64(&mlg.callCount, 1)
		}
	} else {
		for {
			select {
			case <-mlg.ctx.Done():
				spinnerInstance.Stop()
				mlg.prepareStop(mlg.ctx.Err())
				return
			default:
			}
			rawRequest := mlg.caller.BuildReq()
			mlg.caller.Call(rawRequest)
			atomic.AddUint64(&mlg.callCount, 1)
		}
	}

}

func (mlg *myDNSLoaderGenerator) Stop() bool {
	if !atomic.CompareAndSwapUint32(
		&mlg.status, STATUS_RUNNING, STATUS_STOPPING) {
		return false
	}
	globalStatus = STATUS_STOPPING
	mlg.cancelFunc()
	for {
		if atomic.LoadUint32(&mlg.status) == STATUS_STOPPED {
			break
		}
		time.Sleep(time.Microsecond)
	}
	return true
}

func (mlg *myDNSLoaderGenerator) Status() uint32 {
	return atomic.LoadUint32(&mlg.status)
}
func (mlg *myDNSLoaderGenerator) CallCount() uint64 {
	return atomic.LoadUint64(&mlg.callCount)
}

// NewDNSLoaderGenerator will return a new instance of generator
// using param from GeneratorParam
func NewDNSLoaderGenerator(param GeneratorParam) (Generator, error) {
	if err := param.ValidCheck(); err != nil {
		return nil, err
	}
	mlg := &myDNSLoaderGenerator{
		caller:   param.Caller,
		timeout:  param.Timeout,
		qps:      param.QPS,
		duration: param.Duration,
		status:   STATUS_STOPPED,
	}
	mlg.result = make(map[uint8]uint64)
	return mlg, nil
}

// GenTrafficFromConfig function will do traffic generate job
// from configuration
func GenTrafficFromConfig(config *Configuration) {
	dnsclient, err := NewDNSClientWithConfig(config)
	if err != nil {
		log.Panicf("%s", err.Error())
	}
	log.Infoln("config the dns loader success")
	log.Infof("dnsloader server info : server:%s|port:%d",
		dnsclient.Config.Server, dnsclient.Config.Port)
	param := GeneratorParam{
		Caller:   dnsclient,
		Timeout:  1000 * time.Millisecond,
		QPS:      uint32(config.QPS),
		Duration: time.Second * time.Duration(config.Duration),
	}
	log.Infof("initialize load %s", param.Info())
	gen, err := NewDNSLoaderGenerator(param)
	if err != nil {
		log.Panicf("load generator initialization fail :%s", err)
	}
	GloablGenerator = gen
	gen.Start()
}
