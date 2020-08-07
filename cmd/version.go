package cmd

import (
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of Baetyl",
	Long:  `The versions of Baetyl is as follows`,
	Run: func(_ *cobra.Command, _ []string) {
		utils.PrintVersion()
	},
}
