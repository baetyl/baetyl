package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	Args: cobra.MinimumNArgs(1),
}

func Execute() {
	rootCmd.Version = utils.VERSION
	rootCmd.SetVersionTemplate(utils.Version())
	cobra.OnInitialize(initConfig)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr,err)
		os.Exit(1)
	}
}

func initConfig() {
	viper.SetEnvPrefix("baetyl")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
}