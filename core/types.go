package core

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

// RawRequest define the raw request with raw bytes
type RawRequest struct {
	ID  int64
	Req []byte
}

// RawResponse define the raw result from call function
type RawResponse struct {
	ID     int64
	Resp   []byte
	Err    error
	Elapse time.Duration
}

// CallResult define the result struct from call function
type CallResult struct {
	ID     int64
	Req    RawRequest
	Resp   RawResponse
	Code   ReturnCode
	Msg    string
	Elapse time.Duration
}

// Generator define the behavior of loader
type Generator interface {
	Start() bool
	Stop() bool
	Status() uint32
	CallCount() int64
}

// Caller define the behavior of call processor
type Caller interface {
	BuildReq() RawRequest
	Call(req []byte, timeout time.Duration) ([]byte, error)
	CheckResp(rawReq RawRequest, rawResp RawResponse) *CallResult
}

// GeneratorParam will be used to new a generator instance as a param
type GeneratorParam struct {
	Caller        Caller
	Timeout       time.Duration
	QPS           uint32
	Duration      time.Duration
	ResultChannel chan *CallResult
}

func (param *GeneratorParam) ValidCheck() error {
	var errMsgs []string
	if param.Caller == nil {
		errMsgs = append(errMsgs, "Invalid caller!")
	}
	if param.Timeout == 0 {
		errMsgs = append(errMsgs, "Invalid timeout!")
	}
	if param.QPS == 0 {
		errMsgs = append(errMsgs, "Invalid qps(query per second)!")
	}
	if param.Duration == 0 {
		errMsgs = append(errMsgs, "Invalid duration!")
	}
	if param.ResultChannel == nil {
		errMsgs = append(errMsgs, "Invalid result channel!")
	}
	var buf bytes.Buffer
	buf.WriteString("Checking the parameters...")
	if errMsgs != nil {
		errMsg := strings.Join(errMsgs, " ")
		buf.WriteString(fmt.Sprintf("NOT passed! (%s)\n", errMsg))
		log.Panic(buf.String())
		return errors.New(errMsg)
	}
	buf.WriteString(
		fmt.Sprintf("Passed. (timeout=%s, qps=%d, duration=%s)",
			param.Timeout, param.QPS, param.Duration))
	log.Println(buf.String())
	return nil
}
