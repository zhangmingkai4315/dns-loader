package core

import (
	"context"
	"encoding/binary"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/zhangmingkai4315/dns-loader/dns"

	log "github.com/sirupsen/logrus"
	"go.uber.org/ratelimit"
)

// LoadParams will be used to new a loader instance with this param
type LoadParams struct {
	Caller       LoadCaller
	Timeout      time.Duration
	QPS          uint32
	Max          uint64
	ClientNumber int
	Protocol     string
	Duration     time.Duration
}

// Info return the basic info of lodaer params
func (param *LoadParams) Info() string {
	return fmt.Sprintf("current loader[qps=%d, durations=%v,timeout=%v]", param.QPS, param.Duration, param.Timeout)
}

//ValidCheck function
func (param *LoadParams) ValidCheck() error {
	var errMsgs []string
	if param.Caller == nil {
		errMsgs = append(errMsgs, "invalid caller!")
	}
	if param.Timeout == 0 {
		errMsgs = append(errMsgs, "invalid timeout!")
	}
	if param.Protocol != "tcp" && param.Protocol != "udp" {
		errMsgs = append(errMsgs, "invalid protocol [tcp, udp] only")
	}
	if param.QPS <= 0 {
		errMsgs = append(errMsgs, "invalid qps setting")
	}
	if param.Max < 0 {
		errMsgs = append(errMsgs, "invalid max setting")
	}
	if param.Duration == 0 {
		errMsgs = append(errMsgs, "invalid duration time")
	}
	if errMsgs != nil {
		errMsg := strings.Join(errMsgs, " ")
		log.Panicf("check the parameters not passed: %s", errMsg)
	}
	log.Infof("check the parameters success. (timeout=%s, qps=%d, max=%d, duration=%s)",
		param.Timeout, param.QPS, param.Max, param.Duration)
	return nil
}

type dnsLoaderGen struct {
	caller         LoadCaller
	protocolOffset int
	timeout        time.Duration
	qps            uint32
	max            uint64
	status         uint32
	duration       time.Duration
	ctx            context.Context
	cancelFunc     context.CancelFunc
	callCount      uint64
	workers        int
	startTime      time.Time
	result         []map[uint8]uint64
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
	app := GetGlobalAppController()
	app.SetCurrentJobStatus(StatusStart)
	if dlg.qps > 0 {
		interval := time.Duration(1e9 / dlg.qps)
		log.Infof("setting throttle %v", interval)
	}
	atomic.StoreUint32(&dlg.status, StatusRunning)
	app.SetCurrentJobStatus(StatusRunning)
	log.Infoln("create new thread to receive dns data from server")
	dnsclient := dlg.caller.(*DNSClient)
	for i := 0; i < dnsclient.NumConn; i++ {
		go func(index int) {
			buf := make([]byte, 4096) // using small tmo buffer for demonstrating
			for {
				n, err := dnsclient.Conn[index].Read(buf)
				if err != nil || n == 0 {
					log.Errorf("error = %v, n=%d", err, n)
					break
				}
				log.Warnf("read n=%d", n)
				if dlg.protocolOffset != 0 {
					//running in tcp mode
					size := binary.BigEndian.Uint16(buf[:2])
					offset := 0
					for n >= (int(size) + 2 + offset) {
						// multi dns packets in single tcp
						log.Warnf("offset =%d", offset)
						code := buf[3+dlg.protocolOffset] & 0x0f
						dlg.result[index][code] = dlg.result[index][code] + 1
						offset += int(size) + 2
					}
				} else {
					// running in udp
					code := buf[3+dlg.protocolOffset] & 0x0f
					dlg.result[index][code] = dlg.result[index][code] + 1
				}
			}
		}(i)
	}

	var limiter ratelimit.Limiter
	if dlg.qps > 0 {
		limiter = ratelimit.New(int(dlg.qps))
	}
	dlg.startTime = time.Now()
	log.Printf("start send dns packets to server and will stop at %s later", dlg.duration)
	dlg.generatorLoad(limiter)
	return true
}

func (dlg *dnsLoaderGen) prepareStop() {
	log.Printf("prepare to stop load test")
	atomic.StoreUint32(&dlg.status, StatusStopping)
	app := GetGlobalAppController()
	app.SetCurrentJobStatus(StatusStopping)
	time.Sleep(2 * time.Second)
	log.Infoln("doing calculation work")
	runningTime := time.Since(dlg.startTime)
	managerCounter := dlg.CallCount()
	log.WithFields(log.Fields{"result": true}).Infof("total packets sum:%d", managerCounter)
	log.WithFields(log.Fields{"result": true}).Infof("runing time %v", runningTime)
	var globalCounter uint64
	globalStatusCounter := make(map[string]uint64)
	for _, clientResult := range dlg.result {
		for k, v := range clientResult {
			globalCounter = v + globalCounter
			old, ok := globalStatusCounter[dns.DNSRcodeReverse[k]]
			if ok == true {
				globalStatusCounter[dns.DNSRcodeReverse[k]] = old + v
			} else {
				globalStatusCounter[dns.DNSRcodeReverse[k]] = v
			}
		}
	}
	for k, v := range globalStatusCounter {
		log.WithFields(log.Fields{"result": true}).Infof("status %s:%d [%.2f]", k, v, float64(v*100)/float64(dlg.CallCount()))
	}

	var unknown uint64
	if managerCounter > globalCounter {
		unknown = managerCounter - globalCounter
	}
	log.WithFields(log.Fields{"result": true}).Infof("status unknown:%d [%.2f]", unknown, float64(unknown*100)/float64(dlg.CallCount()))
	atomic.StoreUint32(&dlg.status, StatusStopped)
	app.SetCurrentJobStatus(StatusStopped)
	dnsclient := dlg.caller.(*DNSClient)
	for _, client := range dnsclient.Conn {
		client.Close()
	}
	log.Info("stop success!")
}

func (dlg *dnsLoaderGen) generatorLoad(limiter ratelimit.Limiter) {
	app := GetGlobalAppController()
	job := app.JobConfig
	for {
		select {
		case <-dlg.ctx.Done():
			dlg.prepareStop()
			return
		default:
		}
		limiter.Take()
		rawRequest := dlg.caller.BuildReq(job)
		dlg.caller.Call(rawRequest)
		atomic.AddUint64(&dlg.callCount, 1)
		if dlg.max != 0 && dlg.callCount >= dlg.max {
			dlg.prepareStop()
			return
		}
	}
}

func (dlg *dnsLoaderGen) Stop() bool {
	if !atomic.CompareAndSwapUint32(
		&dlg.status, StatusRunning, StatusStopping) {
		return false
	}
	app := GetGlobalAppController()
	app.SetCurrentJobStatus(StatusStopping)
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
	offset := 0
	if param.Protocol == "tcp" {
		offset = 2
	}
	dlg := &dnsLoaderGen{
		caller:         param.Caller,
		timeout:        param.Timeout,
		qps:            param.QPS,
		protocolOffset: offset,
		max:            param.Max,
		duration:       param.Duration,
		status:         StatusStopped,
	}
	for i := 0; i < param.ClientNumber; i++ {
		r := make(map[uint8]uint64)
		dlg.result = append(dlg.result, r)
	}
	return dlg, nil
}

// GenTrafficFromConfig function will do traffic generate job
// from configuration
func GenTrafficFromConfig(appController *AppController) error {
	dnsclient, err := NewDNSClient(appController)
	if err != nil {
		log.Errorf("create dns client fail:%s", err)
		return err
	}
	duration, _ := time.ParseDuration(appController.JobConfig.Duration)
	if err != nil {
		log.Errorf("parse user input duration fail :%s", err)
		return err
	}
	param := LoadParams{
		Caller:       dnsclient,
		Timeout:      1000 * time.Millisecond,
		QPS:          appController.QPS,
		Max:          appController.MaxQuery,
		ClientNumber: appController.ClientNumber,
		Duration:     duration,
		Protocol:     appController.Protocol,
	}
	log.Infof("initialize load %s", param.Info())
	gen, err := NewDNSLoaderGenerator(param)
	if err != nil {
		log.Errorf("load generator initialization fail :%s", err)
		return err
	}
	appController.LoadManager = gen
	gen.Start()
	return nil
}
