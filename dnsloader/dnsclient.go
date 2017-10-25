package dnsloader

import (
	"bytes"
	"github.com/zhangmingkai4315/dns-loader/core"
	"github.com/zhangmingkai4315/go-dns-shooter/dns"
	"log"
	"net"
	"strconv"
	"time"
)

// DNSClient 定义dnsloader发包的配置数据
type DNSClient struct {
	Addr   string
	Port   int
	packet *dns.DNSPacket
	Config *Config
	conn   *net.UDPConn
}

// NewDefaultConfig func return the default configration object
func NewDefaultConfig() (config *Config) {
	return &Config{
		// 使用本地ip地址
		LocalIP: true,
		// 是否源地址固定
		// SourceIPFixed:false,
		// 是否固定域名
		DomainFixed: false,
		// 固定部分的域名
		Domain: "example",
		// 随机域名长度
		DomainRandomLength: 10,
		// 是否查询类型固定
		QueryTypeFixed: true,
		// 固定的查询类型
		QueryType: "A",
	}
}

// NewDNSClientWithDefaultConfig 使用默认的发包参数进行初始化
func NewDNSClientWithDefaultConfig(addr string, port int) (*DNSClient, error) {
	config := NewDefaultConfig()
	dnsclient, err := NewDNSClientWithConfig(addr, port, config)
	if err != nil {
		return nil, err
	}
	return dnsclient, nil
}

// NewDNSClientWithConfig 使用传递的配置发包参数进行初始化
func NewDNSClientWithConfig(addr string, port int, config *Config) (dnsclient *DNSClient, err error) {
	if port > 65535 || port < 0 {
		port = 53
	}

	dnsclient = &DNSClient{Addr: addr, Port: port, Config: config}
	udpAddr, err := net.ResolveUDPAddr("udp", addr+":"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, err
	}
	dnsclient.conn = conn
	log.Println("New DNS client success, start send packet...")
	dnsclient.InitPacket()
	return dnsclient, nil
}

// InitPacket init a packet for dns query data
func (client *DNSClient) InitPacket() {
	client.packet = new(dns.DNSPacket)
	if client.Config.QueryTypeFixed == true {
		client.packet.InitialPacket(client.Config.Domain,
			client.Config.DomainRandomLength,
			GetDNSTypeCodeFromString(client.Config.QueryType))
	} else {
		client.packet.InitialPacket(client.Config.Domain,
			client.Config.DomainRandomLength,
			dns.TypeA)
		client.packet.RandomType = true
	}
	log.Println("Init DNS client configuration success")
}

// BuildReq build new dns request for use later
func (client *DNSClient) BuildReq() core.RawRequest {
	id := time.Now().UnixNano()
	randomDomain := dns.GenRandomDomain(client.Config.DomainRandomLength, client.Config.Domain)
	if _, err := client.packet.UpdateSubDomainToBytes(randomDomain); err != nil {
		log.Printf("%v\n", err)
	}
	rawReq := core.RawRequest{
		ID:  id,
		Req: client.packet.RawByte,
	}
	return rawReq
}

// Call func will be called by schedual each time
func (client *DNSClient) Call(req []byte, timeout time.Duration) ([]byte, error) {

	_, err := client.conn.Write(req)
	if err != nil {
		log.Printf("Send DNS Query Failed:%s", err)
		return nil, err
	}
	readBytes := make([]byte, 512)
	var buffer bytes.Buffer
	_, err = client.conn.Read(readBytes)
	if err != nil {
		return nil, err
	}
	readByte := readBytes[0]
	buffer.WriteByte(readByte)
	return buffer.Bytes(), nil
}

// CheckResp func
func (client *DNSClient) CheckResp(rawReq core.RawRequest, rawResponse core.RawResponse) *core.CallResult {
	var result core.CallResult
	result.ID = rawResponse.ID
	result.Req = rawReq
	result.Resp = rawResponse
	// var sreq ServerReq
	// err := json.Unmarshal(rawReq.Req, &sreq)
	// if err != nil {
	// 	result.Code = core.RET_FORMERR
	// 	return &result
	// }

	// var sresp ServerResp
	// err = json.Unmarshal(rawResponse.Resp, &sresp)
	// if err != nil {
	// 	result.Code = core.RET_RESULT_ERROR
	// 	result.Msg =
	// 		fmt.Sprintf("Incorrectly formatted Resp: %s!\n", string(rawResponse.Resp))
	// 	return &result
	// }
	// if sresp.ID != sreq.ID {
	// 	result.Code = core.RET_CODE_ID_ERROR
	// 	result.Msg =
	// 		fmt.Sprintf("Inconsistent raw id! (%d != %d)\n", rawReq.ID, rawResponse.ID)
	// 	return &result
	// }
	// if sresp.Err != nil {
	// 	result.Code = core.RET_SERVER_ERROR
	// 	result.Msg =
	// 		fmt.Sprintf("Abnormal server: %s!\n", sresp.Err)
	// 	return &result
	// }
	result.Code = core.RET_NO_ERROR
	return &result

}
