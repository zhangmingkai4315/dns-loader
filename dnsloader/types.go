package dnsloader

// import (
// 	"fmt"
// )

// Config define the basic configuration for dns loader
type Config struct {
	// 使用本地IP地址
	LocalIP bool `jsonp:"local_ip"`
	// 是否源地址固定
	// SourceIPFixed bool `json:"source_ip_fixed"`
	// 源地址
	// SourceIP string `json:"source_ip"`
	// 是否固定域名
	DomainFixed bool `json:"domain_fixed"`
	// 固定部分的域名
	Domain string `json:"domain"`
	// 随机域名长度
	DomainRandomLength int `json:"domain_random_length"`
	// 是否查询类型固定
	QueryTypeFixed bool `json:"query_type_fixed"`
	// 固定的查询类型
	QueryType string `json:"query_type"`
}
