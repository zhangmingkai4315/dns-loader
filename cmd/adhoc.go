package cmd

import (
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/zhangmingkai4315/dns-loader/core"
)

var (
	duration     time.Duration
	qps          int
	max          int
	domain       string
	server       string
	port         string
	random       int
	querytype    string
	enableEDNS   bool
	enableDNSSEC bool
)

func init() {
	adhocCmd.Flags().DurationVarP(&duration, "duration", "D", time.Second*60, "send out dns traffic duration")
	adhocCmd.Flags().IntVarP(&qps, "qps", "Q", 100, "qps for dns traffic")
	adhocCmd.Flags().IntVarP(&max, "max", "m", 0, "the maximum number of queries outstanding (set 0 means no limit)")
	adhocCmd.Flags().StringVarP(&domain, "domain", "d", "", "domain name")
	adhocCmd.Flags().StringVarP(&server, "server", "s", "", "dns server ip")
	adhocCmd.Flags().StringVarP(&port, "port", "p", "53", "the server to query")
	adhocCmd.Flags().IntVarP(&random, "random", "r", 5, "prefix random subdomain length")
	adhocCmd.Flags().StringVarP(&querytype, "querytype", "q", "", "random dns query type empty is random type")
	adhocCmd.Flags().BoolVarP(&enableEDNS, "edns", "e", false, "enable edns0")
	adhocCmd.Flags().BoolVarP(&enableDNSSEC, "dnssec", "o", false, "set dnssec ok bit")
}

var adhocCmd = &cobra.Command{
	Use:   "adhoc",
	Short: "Run dnsloader in adhoc mode",
	Long:  `Run dnsloader in adhoc mode using arguments to gen dns packets and quit the process when job done`,
	Run: func(cmd *cobra.Command, args []string) {
		var config *core.Configuration
		config = core.GetGlobalConfig()
		config.Domain = domain
		config.DomainRandomLength = random
		config.QPS = uint32(qps)
		config.MaxQuery = uint64(max)
		config.Duration = duration.String()
		config.Server = server
		config.Port = port
		if enableEDNS == true {
			config.EnableEDNS = "true"
		} else {
			config.EnableEDNS = "false"
		}
		if enableDNSSEC == true {
			config.EnableDNSSEC = "true"
		} else {
			config.EnableDNSSEC = "false"
		}
		config.QueryType = querytype
		if err := config.ValidateJobConfiguration(); err != nil {
			log.Panicf("argument validation error:%s", err)
		}
		core.GenTrafficFromConfig(config)
	},
}
