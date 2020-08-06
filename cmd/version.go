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
	Short: "Print the version number of Baetyl",
	Long:  `All software has versions. This is Baetyl's`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(_ *cobra.Command, _ []string) {
		utils.PrintVersion()
	},
}
