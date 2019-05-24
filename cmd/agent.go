package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zhangmingkai4315/dns-loader/web"
)

var agentHost string
var agentPort string
var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Run dnsloader in agent mode",
	Long:  `Run dnsloader in agent mode, receive job from master and gen dns packets`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("start agent server at %s:%s", agentHost, agentPort)
		web.NewAgentServer(agentHost, agentPort)
		return
	},
}

func init() {
	agentCmd.Flags().StringVar(&agentHost, "host", "0.0.0.0", "ipaddress for start agent")
	agentCmd.Flags().StringVar(&agentPort, "port", "8998", "port to listen")
}
