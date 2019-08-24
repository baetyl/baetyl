package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "baetyl",
	Short: "baetyl " + Version + "\nbaetyl extends cloud computing, data and service seamlessly to edge devices",
	Long:  ``,
}

// Execute execute
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func init() {

}

func initConfig() {

}
