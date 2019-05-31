package core

import (
	// "bytes"
	"math/rand"
	"net"

	log "github.com/sirupsen/logrus"

	"github.com/zhangmingkai4315/dns-loader/dns"
)

// DNSClient hold the loader configuration setting and connection
type DNSClient struct {
	packet  *dns.Packet
	Conn    []net.Conn
	NumConn int
	Offset  int
}

// NewDNSClient create a new DNSClient instance
func NewDNSClient(app *AppController) (dnsclient *DNSClient, err error) {
	dnsclient = &DNSClient{
		Conn:    []net.Conn{},
		NumConn: 0,
	}

	clientNumber := app.JobConfig.ClientNumber
	protocal := app.JobConfig.Protocol
	for i := 0; i < clientNumber; i++ {
		conn, err := net.Dial(protocal, app.Server+":"+app.Port)
		if err != nil {
			return nil, err
		}
		dnsclient.Conn = append(dnsclient.Conn, conn)
	}
	dnsclient.NumConn = clientNumber
	log.Println("new dns loader client success")
	err = dnsclient.InitPacket(app.JobConfig)
	if err != nil {
		return nil, err
	}
	return dnsclient, nil
}

// InitPacket init a packet for dns query data
func (client *DNSClient) InitPacket(job *JobConfig) error {
	enableEDNS := false
	enableDNSSEC := false
	if job.EnableEDNS == "true" {
		enableEDNS = true
	}
	if job.EnableDNSSEC == "true" {
		enableDNSSEC = true
	}

	client.packet = new(dns.Packet)
	client.Offset = 0
	if job.Protocol == "tcp" {
		client.Offset = 2
	}
	if job.QueryType != "" {
		queryTypeCode, err := dns.GetDNSTypeCodeFromString(job.QueryType)
		if err != nil {
			log.Errorf("init packet fail: %s", err.Error())
			return err
		}
		client.packet.InitialPacket(
			job.Protocol,
			job.Domain,
			job.DomainRandomLength,
			queryTypeCode,
			enableEDNS,
			enableDNSSEC,
		)
		return nil
	}

	client.packet.InitialPacket(
		job.Protocol,
		job.Domain,
		job.DomainRandomLength,
		dns.TypeA,
		enableEDNS,
		enableDNSSEC,
	)
	client.packet.RandomType = true

	return nil
}

// BuildReq build new dns request for use later
func (client *DNSClient) BuildReq(job *JobConfig) []byte {
	randomDomain := dns.GenRandomDomain(job.DomainRandomLength, job.Domain)
	if _, err := client.packet.UpdateSubDomainToBytes(randomDomain, client.Offset); err != nil {
		log.Printf("%v\n", err)
	}
	return client.packet.RawByte
}

// Call func will be called by schedual each time
func (client *DNSClient) Call(req []byte) {
	n := rand.Intn(client.NumConn)

	_, err := client.Conn[n].Write(req)
	if err != nil {
		log.Printf("send dns query Failed:%s", err)
		return
	}
}
