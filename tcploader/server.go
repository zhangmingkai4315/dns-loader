package tcploader

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"sync/atomic"
)

func reqHandler(conn net.Conn) {
	var errMsg string
	var sresp ServerResp
	req, err := read(conn, DELIM)
	if err != nil {
		errMsg = fmt.Sprintf("Server: Req Read Error:%s", err)
	} else {
		var sreq ServerReq
		err := json.Unmarshal(req, &sreq)
		if err != nil {
			errMsg = fmt.Sprintf("Server:Req Unmarshal Error:%s", err)
		} else {
			sresp.ID = sreq.ID
			sresp.Result = op(sreq.Operands, sreq.Operator)
		}
	}
	if errMsg != "" {
		sresp.Err = errors.New(errMsg)
	}
	bytes, err := json.Marshal(sresp)
	if err != nil {
		log.Printf("Server: Resp Marshal Error: %s\n", err)
	}
	_, err = write(conn, bytes, DELIM)
	if err != nil {
		log.Printf("Server: Resp Write error: %s", err)
	}

}

type TCPServer struct {
	listener net.Listener
	active   uint32
}

func NewTCPServer() *TCPServer {
	return &TCPServer{}
}

func (server *TCPServer) init(addr string) error {
	if !atomic.CompareAndSwapUint32(&server.active, 0, 1) {
		return nil
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		atomic.StoreUint32(&server.active, 0)
		return err
	}
	server.listener = ln
	return nil
}

func (server *TCPServer) Start(addr string) {
	err := server.init(addr)
	if err != nil {
		return
	}
	for {
		if atomic.LoadUint32(&server.active) != 1 {
			break
		}
		conn, err := server.listener.Accept()
		if err != nil {
			if atomic.LoadUint32(&server.active) == 1 {
				log.Printf("Server: Request Acception Error: %s\n", err)
			} else {
				log.Printf("Server: Broken acception because of closed network connection:%v", err)
			}
			continue
		}
		go reqHandler(conn)
	}
}

func (server *TCPServer) Close() bool {
	if !atomic.CompareAndSwapUint32(&server.active, 1, 0) {
		return false
	}
	server.listener.Close()
	return true
}
