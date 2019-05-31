package dns

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
)

// DNSHeader holds a DNS Header.
type DNSHeader struct {
	ID                 uint16
	Response           bool
	Opcode             int
	Authoritative      bool
	Truncated          bool
	RecursionDesired   bool
	RecursionAvailable bool
	Zero               bool
	AuthenticatedData  bool
	CheckingDisabled   bool
	Rcode              int
}

// RawHeader is the wire format for the DNS packet header.
type RawHeader struct {
	ID                                 uint16
	Bits                               uint16
	Qdcount, Ancount, Nscount, Arcount uint16
}

func packUint16(i uint16, msg []byte, off int) (off1 int, err error) {
	if off+2 > len(msg) {
		return len(msg), errors.New("overflow packing uint16")
	}
	binary.BigEndian.PutUint16(msg[off:], i)
	return off + 2, nil
}

func (dh *RawHeader) pack(msg []byte, off int) (int, error) {
	off, err := packUint16(dh.ID, msg, off)
	if err != nil {
		return off, err
	}
	off, err = packUint16(dh.Bits, msg, off)
	if err != nil {
		return off, err
	}
	off, err = packUint16(dh.Qdcount, msg, off)
	if err != nil {
		return off, err
	}
	off, err = packUint16(dh.Ancount, msg, off)
	if err != nil {
		return off, err
	}
	off, err = packUint16(dh.Nscount, msg, off)
	if err != nil {
		return off, err
	}
	off, err = packUint16(dh.Arcount, msg, off)
	return off, err
}

// Question holds a DNS question. There can be multiple questions in the
// question section of a message. Usually there is just one.
type Question struct {
	Name   string
	Qtype  uint16
	Qclass uint16
}

func (question *Question) pack(msg []byte, offset int, rawpack []byte) (int, error) {
	for i, v := range rawpack {
		msg[offset+i] = v
	}
	offset += len(rawpack)
	offset, err := packUint16(question.Qtype, msg, offset)
	if err != nil {
		return offset, err
	}
	offset, err = packUint16(question.Qclass, msg, offset)
	if err != nil {
		return offset, err
	}
	return offset, nil
}

// Packet holds a DNS packet
type Packet struct {
	Protocol       string
	Header         DNSHeader
	Questions      uint16
	Answers        uint16
	AuthorityRRs   uint16
	AdditionRRs    uint16
	Question       []Question // Holds the RR(s) of the question section.
	RawByte        []byte
	init           bool
	lock           sync.Mutex
	RandomLength   int
	RandomType     bool
	OriginalDomain string
}

// SetQuestion will set the basic dns packet infomation
func (dns *Packet) SetQuestion(name string, dnstype uint16, enableEDNS, enableDNSSEC bool) *Packet {
	dns.Header.ID = GenerateRandomID(true)
	dns.Header.RecursionDesired = true
	log.Infof("enable ends = %v enable dnssec = %v", enableEDNS, enableDNSSEC)
	if enableDNSSEC || enableEDNS {
		dns.Header.AuthenticatedData = true
		dns.AdditionRRs = 1
	}
	dns.Questions = 1
	dns.Question = make([]Question, 1)
	dns.Question[0] = Question{name, dnstype, ClassINET}
	return dns
}

// ToBytes will generate the first raw bytes of the dns packet
func (dns *Packet) ToBytes(edns bool, dnssec bool) (msg []byte, err error) {
	var rawheader RawHeader
	header := dns.Header
	rawheader.ID = header.ID
	var ednsBytes = []byte{0, 0, 41, 16, 0, 0, 0, 0, 0, 0, 0}
	var dnssecBytes = []byte{0, 0, 41, 16, 0, 0, 0, 128, 0, 0, 0}
	rawheader.Bits = uint16(header.Opcode)<<11 | uint16(header.Rcode)
	if header.Response {
		rawheader.Bits |= _QR
	}
	if header.Authoritative {
		rawheader.Bits |= _AA
	}
	if header.Truncated {
		rawheader.Bits |= _TC
	}
	if header.RecursionDesired {
		rawheader.Bits |= _RD
	}
	if header.RecursionAvailable {
		rawheader.Bits |= _RA
	}
	if header.Zero {
		rawheader.Bits |= _Z
	}
	if header.AuthenticatedData {
		rawheader.Bits |= _AD
	}
	if header.CheckingDisabled {
		rawheader.Bits |= _CD
	}
	question := dns.Question
	rawheader.Qdcount = uint16(len(question))
	rawheader.Arcount = dns.AdditionRRs
	offset := 0
	formatName := PackDomainName(FqdnFormat(question[0].Name))
	packLen := 12 + len(formatName) + 4
	msg = make([]byte, packLen)
	offset, err = rawheader.pack(msg, offset)
	if err != nil {
		return nil, err
	}
	offset, err = dns.Question[0].pack(msg, offset, formatName)
	if err != nil {
		return nil, err
	}
	if dnssec == true {
		msg = append(msg, dnssecBytes...)
		offset += 11
	} else if edns == true {
		msg = append(msg, ednsBytes...)
		offset += 11
	}

	if dns.Protocol == "tcp" {
		bs := make([]byte, 2)
		binary.BigEndian.PutUint16(bs, uint16(offset))
		msg = append(bs, msg...)
		offset = offset + 2
	}
	dns.init = true
	dns.RawByte = msg[:offset]
	return msg[:offset], nil
}

// UpdateSubDomainToBytes function update the packet []byte with the new domain name
// and return the new raw data
func (dns *Packet) UpdateSubDomainToBytes(domain string, offset int) (msg []byte, err error) {
	// Get a new ID for packet
	id := GenerateRandomID(true)
	// the first two oct is size of packet in tcp mode
	packUint16(id, dns.RawByte, offset)
	rawByte := dns.RawByte[:]
	if len(rawByte) > 0 && dns.init == true {
		formatName := PackDomainName(FqdnFormat(domain))
		for i, v := range formatName {
			rawByte[offset+12+i] = v
		}
		if dns.RandomType {
			offset := offset + 12 + len(formatName)
			packUint16(GenRandomType(), rawByte, offset)
		}
		return rawByte, nil
	}
	return nil, errors.New("Please call ToBytes() before generate more packet")
}

// GeneratePacket will generate dns packet based user input arguments.
func (dns *Packet) GeneratePacket(server string, total int, timeout int, qps int) uint32 {
	var (
		wg                sync.WaitGroup
		MaxProducerNumber int
		ticker            *time.Ticker
		counter           uint32
		jumpOut           bool
		throttle          chan struct{}
	)
	if server == "" {
		server = DefaultServer
	}
	ticker = time.NewTicker(time.Second)
	if runtime.NumCPU() == 1 {
		MaxProducerNumber = 1
	}
	MaxProducerNumber = int(runtime.NumCPU())
	log.Printf("From main goroutine fork %d sub goroutine for generate\n", MaxProducerNumber)

	wg.Add(MaxProducerNumber)
	if qps != 0 && qps > 0 {
		throttle = make(chan struct{}, qps)
		qpsTicker := time.NewTicker(time.Second)
		go func() {
			for {
				select {
				case <-qpsTicker.C:
					for i := 0; i < qps; i++ {
						throttle <- struct{}{}
					}
				}
			}
		}()
	}
	if timeout != 0 && timeout > 0 {
		timerTimeout := time.NewTimer(time.Second * time.Duration(timeout))
		go func() {
			<-timerTimeout.C
			jumpOut = true
		}()
	}
	go func() {
		for {
			select {
			case <-ticker.C:
				log.Printf("Current goroutine number %d [send:%d query]\n", runtime.NumGoroutine(), atomic.LoadUint32(&counter))
			}
		}
	}()

	length := dns.RandomLength
	domain := dns.OriginalDomain
	offset := 0
	if dns.Protocol == "tcp" {
		offset = 2
	}
	for p := 0; p < MaxProducerNumber; p++ {
		conn, err := net.Dial(dns.Protocol, server)
		log.Printf("Open a connection to dns server[%s]\n", server)

		go func() {
			for {
				data := make([]byte, 1024)
				_, err = conn.Read(data)
				if err != nil {
					fmt.Printf("Fail to read udp message:%v\n", err)
					continue
				}
			}
		}()

		go func() {
			if err != nil {
				fmt.Println(err)
			}
			if total == 0 {
				for {
					if qps != 0 && qps > 0 {
						<-throttle
					}
					randomDomain := GenRandomDomain(length, domain)
					rawByte, err := dns.UpdateSubDomainToBytes(randomDomain, offset)
					if err != nil {
						log.Panicf("%v", err)
					}
					conn.Write(rawByte)
					atomic.AddUint32(&counter, 1)
					if jumpOut == true {
						break
					}
				}
				wg.Done()
			} else {
				eachProduceQueryNum := total / MaxProducerNumber
				for i := 0; i < eachProduceQueryNum; i++ {
					if qps != 0 && qps > 0 {
						<-throttle
					}
					randomDomain := GenRandomDomain(length, domain)
					dns.lock.Lock()
					if _, err := dns.UpdateSubDomainToBytes(randomDomain, offset); err != nil {
						log.Panicf("%v", err)
					}
					conn.Write(dns.RawByte)
					dns.lock.Unlock()
					atomic.AddUint32(&counter, 1)
					if jumpOut == true {
						break
					}
				}
				wg.Done()
			}
		}()
	}
	wg.Wait()
	return counter
}

// InitialPacket initial the basic setup
func (dns *Packet) InitialPacket(
	protocol string,
	domain string,
	length int,
	queryType uint16,
	enableEDNS, enableDNSSEC bool,
) {
	log.Infof("dns packet info :[protocol=%s, domain=%s,length=%d,type=%d]", protocol, domain, length, queryType)
	dns.Protocol = protocol
	dns.SetQuestion(FqdnFormat(GenRandomDomain(length, domain)), queryType, enableEDNS, enableDNSSEC)
	dns.ToBytes(enableEDNS, enableDNSSEC)
	dns.RandomLength = length
	dns.OriginalDomain = domain
}
