package core

import (
	// "bytes"
	"net"

	log "github.com/sirupsen/logrus"

	"github.com/zhangmingkai4315/dns-loader/dns"
)

// DNSClient hold the loader configuration setting and connection
type DNSClient struct {
	packet *dns.DNSPacket
	Config *Configuration
	Conn   net.Conn
}

// NewDNSClientWithConfig create a new DNSClient instance
func NewDNSClientWithConfig(config *Configuration) (dnsclient *DNSClient, err error) {
	dnsclient = &DNSClient{Config: config}
	conn, err := net.Dial("udp", config.Server+":"+config.Port)
	if err != nil {
		return nil, err
	}
	dnsclient.Conn = conn
	log.Println("new dns loader client success")
	err = dnsclient.InitPacket()
	if err != nil {
		return nil, err
	}
	return dnsclient, nil
}

// InitPacket init a packet for dns query data
func (client *DNSClient) InitPacket() error {
	client.packet = new(dns.DNSPacket)
	// client.packet.InitialPacket(domain, randomlen, dns.TypeA)
	if client.Config.QueryType != "" {
		queryTypeCode, err := dns.GetDNSTypeCodeFromString(client.Config.QueryType)
		if err != nil {
			log.Errorf("init packet fail: %s", err.Error())
			return err
		}
		client.packet.InitialPacket(client.Config.Domain,
			client.Config.DomainRandomLength, queryTypeCode)
		return nil
	}
	client.packet.InitialPacket(client.Config.Domain,
		client.Config.DomainRandomLength,
		dns.TypeA)
	client.packet.RandomType = true
	return nil
}

// BuildReq build new dns request for use later
func (client *DNSClient) BuildReq() []byte {
	randomDomain := dns.GenRandomDomain(client.Config.DomainRandomLength, client.Config.Domain)
	if _, err := client.packet.UpdateSubDomainToBytes(randomDomain); err != nil {
		log.Printf("%v\n", err)
	}
	return client.packet.RawByte
}

// Call func will be called by schedual each time
func (client *DNSClient) Call(req []byte) {
	_, err := client.Conn.Write(req)
	if err != nil {
		log.Printf("send dns query Failed:%s", err)
		return
	}
}
