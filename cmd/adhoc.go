package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/zhangmingkai4315/dns-loader/core"
)

var (
	duration   int
	qps        int
	domain     string
	server     string
	port       int
	random     int
	randomType bool
	querytype  string
)

func init() {
	adhocCmd.PersistentFlags().IntVarP(&duration, "duration", "D", 60, "send out dns traffic duration")
	adhocCmd.PersistentFlags().IntVarP(&qps, "qps", "Q", 100, "qps for dns traffic")
	adhocCmd.PersistentFlags().StringVarP(&domain, "domain", "d", "", "domain name")
	adhocCmd.PersistentFlags().StringVarP(&server, "server", "s", "", "dns server ip")
	adhocCmd.PersistentFlags().IntVarP(&port, "port", "p", 53, "dns server port")
	adhocCmd.PersistentFlags().IntVarP(&random, "random", "r", 5, "prefix random subdomain length")
	adhocCmd.PersistentFlags().BoolVarP(&randomType, "randomtype", "R", false, "random dns query type")
	adhocCmd.PersistentFlags().StringVarP(&querytype, "querytype", "q", "A", "random dns query type")
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
		config.QPS = qps
		config.Duration = duration
		config.QueryTypeFixed = randomType
		config.Server = server
		config.Port = port
		config.QueryType = querytype
		if err := config.Valid(); err != nil {
			log.Panicf("argument validation error:%s", err)
		}
		core.GenTrafficFromConfig(config)
	},
}
