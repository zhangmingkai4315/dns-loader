package cmd

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zhangmingkai4315/dns-loader/core"
	"github.com/zhangmingkai4315/dns-loader/web"
)

var masterConfigFile, dbFile string
var masterCmd = &cobra.Command{
	Use:   "master",
	Short: "Run dnsloader in master mode",
	Long:  `Run dnsloader in master mode, using website to config the job and send commands to slave and control it`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Infoln("start dnsloader in master mode")
		config := initMasterMode(masterConfigFile, dbFile)
		log.Infof("load config file from %s success", masterConfigFile)
		log.Printf("start web for dnsloader admin :%s", config.HTTPServer)
		log.Printf("default User:%s Password:%s", config.User, config.Password)
		err := web.NewServer()
		if err != nil {
			log.Printf("start web server fail: %s", err)
		}
		return
	},
}

func init() {
	masterCmd.PersistentFlags().StringVar(&masterConfigFile, "config", "", "config file (default is $HOME/config.ini)")
	masterCmd.PersistentFlags().StringVar(&dbFile, "dbfile", "app.db", "database file for dns loader app")
}

func initMasterMode(cfgFile string, dbfile string) *core.Configuration {
	if cfgFile == "" {
		// Find current application directory.
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			log.Fatal(err)
		}
		cfgFile = filepath.Join(dir, "config.ini")
	}
	var err error
	config, err := core.NewConfigurationFromFile(cfgFile)
	if err != nil {
		log.Fatalf("load config file %s error: %s", cfgFile, err)
	}
	err = core.NewDatabaseFromFile(dbfile)
	if err != nil {
		log.Fatalf("create database connection error: %s", err.Error())
	}
	log.Infof("using sqlite as app database filename is %s", dbfile)
	return config
}
