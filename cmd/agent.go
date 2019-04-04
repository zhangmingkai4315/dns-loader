package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zhangmingkai4315/dns-loader/web"
)

var agentConfigFile string
var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Run dnsloader in agent mode",
	Long:  `Run dnsloader in agent mode, receive job from master and gen dns packets`,
	Run: func(cmd *cobra.Command, args []string) {
		config := initConfig(agentConfigFile)
		if config.AgentPort == 0 || config.ControlMaster == "" {
			log.Fatalln("agent port and master ip must given")
		}
		log.Printf("start agent server listen on %d for master:%s connect", config.AgentPort, config.ControlMaster)
		web.NewAgentServer(config)
		return
	},
}

func init() {
	agentCmd.PersistentFlags().StringVar(&agentConfigFile, "config", "-c", "config file (default is $HOME/config.ini)")
}
