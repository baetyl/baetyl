package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/baetyl/baetyl/v2/program"
)

func init() {
	rootCmd.AddCommand(programCmd)
}

var programCmd = &cobra.Command{
	Use:   "program",
	Short: "Run a program of Baetyl",
	Long:  `Baetyl loads program's configuration from program_service.yml, then runs and waits the program to stop.`,
	Run: func(_ *cobra.Command, args []string) {
		var wd string
		if runtime.GOOS == "windows" {
			wd = args[0]
		}
		if err := program.Run(wd); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	},
}
