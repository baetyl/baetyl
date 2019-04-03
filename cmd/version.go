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
	Short: "Show the OpenEdge version information",
	Long:  ``,
	Run:   version,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func version(cmd *cobra.Command, args []string) {
	log.Printf("\nOpenEdge version %s\n%s\n\n",
		Version,
		GoVersion)
}
