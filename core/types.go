package core

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// Generator define the behavior of loader
type Generator interface {
	Start() bool
	Stop() bool
	Status() uint32
	CallCount() uint64
}

// Caller define the behavior of call processor
type Caller interface {
	BuildReq() []byte
	Call(req []byte)
}

// GeneratorParam will be used to new a generator instance as a param
type GeneratorParam struct {
	Caller   Caller
	Timeout  time.Duration
	QPS      uint32
	Duration time.Duration
}

// Info return the basic info of gernerator params
func (param *GeneratorParam) Info() string {
	return fmt.Sprintf("gernerator[qps=%d, durations=%v,timeout=%v]", param.QPS, param.Duration, param.Timeout)
}

//ValidCheck function
func (param *GeneratorParam) ValidCheck() error {
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
	if param.Duration == 0 {
		errMsgs = append(errMsgs, "invalid duration time")
	}
	if errMsgs != nil {
		errMsg := strings.Join(errMsgs, " ")
		log.Panicf("check the parameters not passed: %s", errMsg)
	}
	log.Infof("check the parameters success. (timeout=%s, qps=%d, duration=%s)",
		param.Timeout, param.QPS, param.Duration)
	return nil
}
