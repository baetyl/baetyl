package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Compile parameter
var (
	Version  string
	Revision string
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
	fmt.Printf("Version:      %s\nGit revision: %s\nGo version:   %s\n\n", Version, Revision, runtime.Version())
}
