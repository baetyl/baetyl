package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/baetyl/baetyl/program"
)

func init() {
	rootCmd.AddCommand(programCmd)
}

var programCmd = &cobra.Command{
	Use:   "program",
	Short: "Run a program of Baetyl",
	Long:  `Baetyl loads program's configuration from program_service.yml, then runs and waits the program to stop.`,
	Run: func(_ *cobra.Command, _ []string) {
		if err := program.Run(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	},
}
