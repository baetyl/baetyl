package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

// Compile parameter
var (
	Version   string
	GoVersion string
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "show the version of baetyl",
	Long:  ``,
	Run:   version,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func version(cmd *cobra.Command, args []string) {
	log.Printf("\nbaetyl version %s\n%s\n\n", Version, GoVersion)
}
