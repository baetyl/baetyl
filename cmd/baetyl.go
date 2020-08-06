package cmd

import (
	"fmt"
	"os"

	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "baetyl",
	Short: "Baetyl Command",
	Long: `
	o--o    O  o--o o-O-o o   o o    
	|   |  / \ |      |    \ /  |    
	O--o  o---oO-o    |     O   |    
	|   | |   ||      |     |   |    
	o--o  o   oo--o   o     o   O---o
									 
Extend cloud computing, data and service seamlessly to edge devices.`,
}

func Execute() {
	rootCmd.Version = utils.VERSION
	rootCmd.SetVersionTemplate(utils.Version())
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
