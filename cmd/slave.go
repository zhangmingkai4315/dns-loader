package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zhangmingkai4315/dns-loader/web"
)

var slaveConfigFile string
var slaveCmd = &cobra.Command{
	Use:   "slave",
	Short: "Run dnsloader in slave mode",
	Long:  `Run dnsloader in slave mode, receive job from master and gen dns packets`,
	Run: func(cmd *cobra.Command, args []string) {
		config := initConfig(slaveConfigFile)
		if config.AgentPort == 0 || config.ControlMaster == "" {
			log.Fatalln("agent port and master ip must given")
		}
		log.Printf("start agent server listen on %d for master:%s:%s connect", config.AgentPort, config.ControlMaster)
		web.NewAgentServer(config)
		return
	},
}

func init() {
	slaveCmd.PersistentFlags().StringVar(&slaveConfigFile, "config", "-c", "config file (default is $HOME/config.ini)")
}
