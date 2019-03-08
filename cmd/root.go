package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "openedge",
	Short: "OpenEdge, extend cloud computing, data and service seamlessly to edge devices",
	Long:  `OpenEdge provides an open framework, which allows access to any protocol through a variety of networks, and allows any application to run on multiple systems.`,
}

// Execute execute
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {

}

func initConfig() {

}
