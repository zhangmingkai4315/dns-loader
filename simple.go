package main

import (
    "fmt"
    "net"
	"os"
	"log"
	"github.com/zhangmingkai4315/go-dns-shooter/dns"
    "time"
)
var displayReq bool
var pkts int
var bytes int
var dest string


func main() {
    dest = "172.17.0.2:53"
	packet := new(dns.DNSPacket)
	packet.InitialPacket("test",10,dns.TypeA)
	packet.RandomType = true
	
	log.Println("Init DNS client configuration success")
    go displayTimerProc()

	conn, err := net.Dial("udp", dest)
    if err != nil {
        fmt.Printf("Dial() failed. Error %s\n", err)
        os.Exit(1)
    }
    defer conn.Close()

    pkts = 0
    bytes = 0

    for {
        n, err := conn.Write(packet.RawByte)
        if err != nil {
            //fmt.Printf("Write failed! Error: %s\n", err)
            // os.Exit(1)
        }
        pkts++
        bytes += n
    }
}

func displayTimerProc() {
    for {
        displayReq = true
        time.Sleep(time.Second)
        fmt.Printf("Pkts %d, Bytes %d, rate %d mbps\n",
            pkts, bytes, bytes*8/(1000*1000))
        pkts = 0
        bytes = 0
    }
}