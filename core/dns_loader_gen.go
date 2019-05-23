package core

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/zhangmingkai4315/dns-loader/dns"

	log "github.com/sirupsen/logrus"
	"go.uber.org/ratelimit"
)

//GloablGenerator define global control object
var GloablGenerator Generator

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
	if currentStatus != StatusStopped {
		return false
	}
	atomic.StoreUint32(&mlg.status, StatusStart)
	config := GetGlobalConfig()
	config.SetCurrentJobStatus(StatusStart)
	if mlg.qps > 0 {
		interval := time.Duration(1e9 / mlg.qps)
		log.Infof("setting throttle %v", interval)
	}
	atomic.StoreUint32(&mlg.status, StatusRunning)
	config.SetCurrentJobStatus(StatusRunning)
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
	log.Printf("start push packets to dns server and will stop at %s later...", mlg.duration)
	mlg.generatorLoad(limiter)
	return true
}

func (mlg *myDNSLoaderGenerator) prepareStop(err error) {
	log.Printf("prepare to stop load test [%s]", err)
	atomic.StoreUint32(&mlg.status, StatusStopping)
	config := GetGlobalConfig()
	config.SetCurrentJobStatus(StatusStopping)
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
	var unknown uint64
	if managerCounter > counter {
		unknown = managerCounter - counter
	}
	log.WithFields(log.Fields{"result": true}).Infof("status unknown:%d [%.2f]", unknown, float64(unknown*100)/float64(mlg.CallCount()))
	atomic.StoreUint32(&mlg.status, StatusStopped)
	config.SetCurrentJobStatus(StatusStopped)
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

func (mlg *myDNSLoaderGenerator) generatorLoad(limiter ratelimit.Limiter) {
	for {
		select {
		case <-mlg.ctx.Done():
			mlg.prepareStop(mlg.ctx.Err())
			return
		default:
		}
		limiter.Take()
		rawRequest := mlg.caller.BuildReq()
		mlg.caller.Call(rawRequest)
		atomic.AddUint64(&mlg.callCount, 1)
	}
}

func (mlg *myDNSLoaderGenerator) Stop() bool {
	if !atomic.CompareAndSwapUint32(
		&mlg.status, StatusRunning, StatusStopping) {
		return false
	}
	config := GetGlobalConfig()
	config.SetCurrentJobStatus(StatusStopping)
	mlg.cancelFunc()
	for {
		if atomic.LoadUint32(&mlg.status) == StatusStopped {
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
		status:   StatusStopped,
	}
	mlg.result = make(map[uint8]uint64)
	return mlg, nil
}

// GenTrafficFromConfig function will do traffic generate job
// from configuration
func GenTrafficFromConfig(config *Configuration) error {
	dnsclient, err := NewDNSClientWithConfig(config)
	if err != nil {
		log.Errorf("create dns client fail:%s", err)
		return err
	}
	duration, _ := time.ParseDuration(config.JobConfig.Duration)
	if err != nil {
		log.Errorf("parse user input duration fail :%s", err)
		return err
	}
	param := GeneratorParam{
		Caller:   dnsclient,
		Timeout:  1000 * time.Millisecond,
		QPS:      uint32(config.QPS),
		Duration: duration,
	}
	log.Infof("initialize load %s", param.Info())
	gen, err := NewDNSLoaderGenerator(param)
	if err != nil {
		log.Errorf("load generator initialization fail :%s", err)
		return err
	}
	GloablGenerator = gen
	gen.Start()
	return nil
}
