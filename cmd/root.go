package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dnsloader",
	Short: "dnsloader is a fast dns packets generator and easy to scale in multiple server",
	Long:  "Description: dnsloader is a fast and flexible dns packets generator build with golang, complete documentation is available at https://github.com/zhangmingkai4315/dns-loader",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(masterCmd)
	rootCmd.AddCommand(agentCmd)
	rootCmd.AddCommand(adhocCmd)
	rootCmd.AddCommand(versionCmd)
}

//Execute for cobra root cmd
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
