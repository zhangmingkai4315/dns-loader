package main

import (
	core "github.com/zhangmingkai4315/dns-loader/core"
	tcploader "github.com/zhangmingkai4315/dns-loader/tcploader"
	"log"
	"time"
)

func main() {
	server := tcploader.NewTCPServer()
	defer server.Close()
	serverAddr := "127.0.0.1:9999"
	log.Printf("Start server[%s]\n", serverAddr)
	go server.Start(serverAddr)
	param := core.GeneratorParam{
		Caller:        tcploader.NewTCPClient(serverAddr),
		Timeout:       50 * time.Millisecond,
		QPS:           uint32(1000),
		Duration:      100 * time.Second,
		ResultChannel: make(chan *core.CallResult, 50),
	}
	log.Printf("Initialize load %+v", param)
	gen, err := core.NewLoaderGenerator(param)
	if err != nil {
		log.Panicf("Load generator initialization fail :%s", err)
	}
	log.Println("Start load generator")
	gen.Start()
	log.Println("Counting for result")
	countMap := make(map[core.ReturnCode]int)
	count := 0
	for r := range param.ResultChannel {
		countMap[r.Code] = countMap[r.Code] + 1
		count++
	}

	var total int

	for k, v := range countMap {
		codePlain := core.GetDetailInfo(k)
		log.Printf("Code plain: %s (%d), Count: %d.\n",
			codePlain, k, v)
		total += v
	}
	log.Printf("Total:%d\n", total)
	successCount := countMap[core.RET_NO_ERROR]
	tps := float64(successCount) / float64(param.Duration/1e9)
	log.Printf("Loads per second: %d; Treatments per second: %f.\n", param.QPS, tps)
}
