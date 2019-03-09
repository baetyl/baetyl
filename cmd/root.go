package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "openedge",
	Short: "OpenEdge, extend cloud computing, data and service seamlessly to edge devices",
	Long:  ``,
}

// Execute execute
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
}

func init() {

}

func initConfig() {

}
