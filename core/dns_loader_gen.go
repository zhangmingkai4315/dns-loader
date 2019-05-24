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
var GloablGenerator LoadManager

type dnsLoaderGen struct {
	caller     LoadCaller
	timeout    time.Duration
	qps        uint32
	maxquery   uint64
	status     uint32
	duration   time.Duration
	ctx        context.Context
	cancelFunc context.CancelFunc
	callCount  uint64
	workers    int
	startTime  time.Time
	result     map[uint8]uint64
}

func (dlg *dnsLoaderGen) Start() bool {
	log.Info("prepare dns loader generator")
	dlg.ctx, dlg.cancelFunc = context.WithTimeout(context.Background(), dlg.duration)
	dlg.callCount = 0
	currentStatus := dlg.Status()
	if currentStatus != StatusStopped {
		return false
	}
	atomic.StoreUint32(&dlg.status, StatusStart)
	config := GetGlobalConfig()
	config.SetCurrentJobStatus(StatusStart)
	if dlg.qps > 0 {
		interval := time.Duration(1e9 / dlg.qps)
		log.Infof("setting throttle %v", interval)
	}
	atomic.StoreUint32(&dlg.status, StatusRunning)
	config.SetCurrentJobStatus(StatusRunning)
	log.Infoln("create new thread to receive dns data from server")
	go func() {
		// recive data from connections
		b := make([]byte, 4)
		dnsclient := dlg.caller.(*DNSClient)
		for {
			n, err := dnsclient.Conn.Read(b)
			if err == nil && n > 0 {
				code := b[3] & 0x0f
				dlg.result[code] = dlg.result[code] + 1
			}
		}
	}()

	var limiter ratelimit.Limiter
	if dlg.qps > 0 {
		limiter = ratelimit.New(int(dlg.qps))
	}
	dlg.startTime = time.Now()
	log.Printf("start push packets to dns server and will stop at %s later", dlg.duration)
	dlg.generatorLoad(limiter)
	return true
}

func (dlg *dnsLoaderGen) prepareStop(err error) {
	log.Printf("prepare to stop load test [%s]", err)
	atomic.StoreUint32(&dlg.status, StatusStopping)
	config := GetGlobalConfig()
	config.SetCurrentJobStatus(StatusStopping)
	log.Infoln("doing calculation work")
	runningTime := time.Since(dlg.startTime)
	managerCounter := dlg.CallCount()
	log.WithFields(log.Fields{"result": true}).Infof("total packets sum:%d", managerCounter)
	log.WithFields(log.Fields{"result": true}).Infof("runing time %v", runningTime)
	var counter uint64
	for k, v := range dlg.result {
		counter = v + counter
		log.WithFields(log.Fields{"result": true}).Infof("status %s:%d [%.2f]", dns.DNSRcodeReverse[k], v, float64(v*100)/float64(dlg.CallCount()))
	}
	var unknown uint64
	if managerCounter > counter {
		unknown = managerCounter - counter
	}
	log.WithFields(log.Fields{"result": true}).Infof("status unknown:%d [%.2f]", unknown, float64(unknown*100)/float64(dlg.CallCount()))
	atomic.StoreUint32(&dlg.status, StatusStopped)
	config.SetCurrentJobStatus(StatusStopped)
	log.Info("stop success!")
}

func (dlg *dnsLoaderGen) generatorLoad(limiter ratelimit.Limiter) {
	for {
		select {
		case <-dlg.ctx.Done():
			dlg.prepareStop(dlg.ctx.Err())
			return
		default:
		}
		limiter.Take()
		rawRequest := dlg.caller.BuildReq()
		dlg.caller.Call(rawRequest)
		atomic.AddUint64(&dlg.callCount, 1)
	}
}

func (dlg *dnsLoaderGen) Stop() bool {
	if !atomic.CompareAndSwapUint32(
		&dlg.status, StatusRunning, StatusStopping) {
		return false
	}
	config := GetGlobalConfig()
	config.SetCurrentJobStatus(StatusStopping)
	dlg.cancelFunc()
	for {
		if atomic.LoadUint32(&dlg.status) == StatusStopped {
			break
		}
		time.Sleep(time.Microsecond)
	}
	return true
}

func (dlg *dnsLoaderGen) Status() uint32 {
	return atomic.LoadUint32(&dlg.status)
}
func (dlg *dnsLoaderGen) CallCount() uint64 {
	return atomic.LoadUint64(&dlg.callCount)
}

// NewDNSLoaderGenerator will return a new instance of generator
// using param from GeneratorParam
func NewDNSLoaderGenerator(param LoadParams) (LoadManager, error) {
	if err := param.ValidCheck(); err != nil {
		return nil, err
	}
	dlg := &dnsLoaderGen{
		caller:   param.Caller,
		timeout:  param.Timeout,
		qps:      param.QPS,
		duration: param.Duration,
		status:   StatusStopped,
	}
	dlg.result = make(map[uint8]uint64)
	return dlg, nil
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
	param := LoadParams{
		Caller:   dnsclient,
		Timeout:  1000 * time.Millisecond,
		QPS:      config.QPS,
		Max:      config.MaxQuery,
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
