package cmd

import (
	"fmt"
	"log"
	"runtime"

	"github.com/spf13/cobra"
)

// Compile parameter
var (
	Version   string
	Revision  string
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
	version := fmt.Sprintf("Version:      %s\n", Version)
	version += fmt.Sprintf("Git revision: %s\n", Revision)
	version += fmt.Sprintf("GO version:   %s\n", runtime.Version())
	log.Printf("\n%s", version)
}
