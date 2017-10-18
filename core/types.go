package core

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

type RawRequest struct {
	ID  int64
	Req []byte
}
type RawResponse struct {
	ID     int64
	Resp   []byte
	Err    error
	Elapse time.Duration
}

type CallResult struct {
	ID     int64
	Req    RawRequest
	Resq   RawResponse
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

type GeneratorParam struct {
	Caller        Caller
	Timeout       time.Duration
	Qps           uint32
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
	if param.Qps == 0 {
		errMsgs = append(errMsgs, "Invalid qps(query per second)!")
	}
	if param.Duration == 0 {
		errMsgs = append(errMsgs, "Invalid duration!")
	}
	if param.ResultChannel == nil {
		errMsgs = append(errMsgs, "Invalid result channel!")
	}
	var buf bytes.Buffer
	buf.WriteString("Checking the parameters...\n")
	if errMsgs != nil {
		errMsg := strings.Join(errMsgs, " ")
		buf.WriteString(fmt.Sprintf("NOT passed! (%s)\n", errMsg))
		log.Panic(buf.String())
		return errors.New(errMsg)
	}
	buf.WriteString(
		fmt.Sprintf("Passed. (timeout=%s, qps=%d, duration=%s)\n",
			param.Timeout, param.Qps, param.Duration))
	log.Println(buf.String())
	return nil
}
