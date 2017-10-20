package tcploader

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	core "github.com/zhangmingkai4315/dns-loader/core"
	"math/rand"
	"net"
	"time"
)

type TCPClient struct {
	addr string
}

func NewTCPClient(addr string) *TCPClient {
	return &TCPClient{addr: addr}
}

func (loader *TCPClient) BuildReq() core.RawRequest {
	id := time.Now().UnixNano()
	serverReq := ServerReq{
		ID: id,
		Operands: []int{
			int(rand.Int31n(1000) + 1),
			int(rand.Int31n(1000) + 1),
		},
		Operator: func() string {
			return operators[rand.Int31n(100)%4]
		}(),
	}
	bytes, err := json.Marshal(serverReq)
	if err != nil {
		panic(err)
	}
	rawReq := core.RawRequest{
		ID:  id,
		Req: bytes,
	}
	return rawReq
}

func (loader *TCPClient) Call(req []byte, timeout time.Duration) ([]byte, error) {
	conn, err := net.DialTimeout("tcp", loader.addr, timeout)
	if err != nil {
		return nil, err
	}
	_, err = write(conn, req, DELIM)
	if err != nil {
		return nil, err
	}
	return read(conn, DELIM)
}

func read(conn net.Conn, delim byte) ([]byte, error) {
	readBytes := make([]byte, 1)
	var buffer bytes.Buffer
	for {
		_, err := conn.Read(readBytes)
		if err != nil {
			return nil, err
		}
		readByte := readBytes[0]
		if readByte == delim {
			break
		}
		buffer.WriteByte(readByte)
	}
	return buffer.Bytes(), nil
}

func write(conn net.Conn, content []byte, delim byte) (int, error) {
	write := bufio.NewWriter(conn)
	n, err := write.Write(content)
	if err == nil {
		write.WriteByte(delim)
	}
	if err == nil {
		err = write.Flush()
	}
	return n, err
}

func (loader *TCPClient) CheckResp(
	rawReq core.RawRequest,
	rawResponse core.RawResponse,
) *core.CallResult {
	var result core.CallResult
	result.ID = rawResponse.ID
	result.Req = rawReq
	result.Resp = rawResponse

	var sreq ServerReq
	err := json.Unmarshal(rawReq.Req, &sreq)
	if err != nil {
		result.Code = core.RET_FORMERR
		return &result
	}

	var sresp ServerResp
	err = json.Unmarshal(rawResponse.Resp, &sresp)
	if err != nil {
		result.Code = core.RET_RESULT_ERROR
		result.Msg =
			fmt.Sprintf("Incorrectly formatted Resp: %s!\n", string(rawResponse.Resp))
		return &result
	}
	if sresp.ID != sreq.ID {
		result.Code = core.RET_CODE_ID_ERROR
		result.Msg =
			fmt.Sprintf("Inconsistent raw id! (%d != %d)\n", rawReq.ID, rawResponse.ID)
		return &result
	}
	if sresp.Err != nil {
		result.Code = core.RET_SERVER_ERROR
		result.Msg =
			fmt.Sprintf("Abnormal server: %s!\n", sresp.Err)
		return &result
	}
	result.Code = core.RET_NO_ERROR
	return &result

}
