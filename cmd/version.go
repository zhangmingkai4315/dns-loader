package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const version = "v1.2.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version of dnsloader",
	Long:  `All software has versions. This is dnsloader's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version)
	},
}
