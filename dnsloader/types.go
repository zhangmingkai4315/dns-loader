package dnsloader

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// Config define the basic configuration for dns loader
type Config struct {
	// 使用本地IP地址
	LocalIP bool `jsonp:"local_ip"`
	// 是否源地址固定
	// SourceIPFixed bool `json:"source_ip_fixed"`
	// 源地址
	// SourceIP string `json:"source_ip"`
	// 是否固定域名
	DomainFixed bool `json:"domain_fixed"`
	// 固定部分的域名
	Domain string `json:"domain"`
	// 随机域名长度
	DomainRandomLength int `json:"domain_random_length"`
	// 是否查询类型固定
	QueryTypeFixed bool `json:"query_type_fixed"`
	// 固定的查询类型
	QueryType string `json:"query_type"`
}

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
	if param.QPS < 0 {
		errMsgs = append(errMsgs, "invalid qps (dns query per second)!")
	}
	if param.Duration == 0 {
		errMsgs = append(errMsgs, "invalid duration!")
	}
	log.Infoln("checking the parameters")
	if errMsgs != nil {
		errMsg := strings.Join(errMsgs, " ")
		log.Panicf("check the parameters not passed: %s", errMsg)
	}
	log.Infof("check the parameters success. (timeout=%s, qps=%d, duration=%s)",
		param.Timeout, param.QPS, param.Duration)
	return nil
}
