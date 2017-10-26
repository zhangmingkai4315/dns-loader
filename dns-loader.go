package main

import (
	core "github.com/zhangmingkai4315/dns-loader/core"
	loader "github.com/zhangmingkai4315/dns-loader/dnsloader"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"
)

func main() {
	go func() {
		log.Println("Start performace monitor on port 8080")
		err := http.ListenAndServe("localhost:8080", http.DefaultServeMux)
		if err != nil {
			log.Println("Start performance monitor fail")
		}
	}()
	// Start docker container first
	server := "172.17.0.2"
	port := 53
	dnsclient, err := loader.NewDNSClientWithDefaultConfig(server, port)
	if err != nil {
		log.Panicf("%s", err.Error())
	}
	log.Println("Config the dns loader done")
	log.Printf("Current configuration for dns loader is server:%s|port:%d\nConfigDetail:%+v\n",
		dnsclient.Addr, dnsclient.Port, dnsclient.Config)

	param := core.GeneratorParam{
		Caller:        dnsclient,
		Timeout:       1000 * time.Millisecond,
		QPS:           uint32(100000),
		Duration:      200 * time.Second,
		ResultChannel: make(chan *core.CallResult, 2000),
	}
	log.Printf("Initialize load %+v", param)
	gen, err := loader.NewDNSLoaderGenerator(param)
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
