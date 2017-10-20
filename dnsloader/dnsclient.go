package dnsloader

import (
// "github.com/zhangmingkai4315/dns-loader/core"
)

type DNSClient struct {
	addr   string
	port   int
	config Config
}

func NewDNSClient(addr string, port int) *DNSClient {
	if port > 65535 || port < 0 {
		port = 53
	}
	return &DNSClient{addr: addr, port: port}
}

// func (loader *DNSClient) BuildReq() core.RawRequest {
// 	id := 0
// 	serverReq := buildDNSPacket(loader.config)
// }
