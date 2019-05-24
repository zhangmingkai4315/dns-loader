package core

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// LoadManager define the behavior of dns benchmark loader
type LoadManager interface {
	Start() bool
	Stop() bool
	Status() uint32
	CallCount() uint64
}

// LoadCaller define the behavior of call processor
type LoadCaller interface {
	BuildReq() []byte
	Call(req []byte)
}

// LoadParams will be used to new a loader instance with this param
type LoadParams struct {
	Caller   LoadCaller
	Timeout  time.Duration
	QPS      uint32
	Max      uint64
	Duration time.Duration
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
