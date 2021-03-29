package cmd

import (
	"github.com/spf13/cobra"

	_ "github.com/baetyl/baetyl/v2/ami"
	"github.com/baetyl/baetyl/v2/initz"
)

const (
	HookNameStartInitService = "startInitService"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Run init program of Baetyl",
	Long:  `Baetyl runs the init program to sync with cloud and start core service.`,
	Run: func(_ *cobra.Command, _ []string) {
		// actual func is initz.StartInitService()
		Hooks[HookNameStartInitService].(initz.StartInitServiceFunc)()
	},
}
