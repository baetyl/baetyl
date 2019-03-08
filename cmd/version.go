package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version   string
	BuildTime string
	GoVersion string
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the OpenEdge version information",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("OpenEdge version %s\nbuild time %s\n%s\n\n",
			Version,
			BuildTime,
			GoVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
