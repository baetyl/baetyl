package cmd

import (
	"github.com/spf13/cobra"

	"github.com/baetyl/baetyl/v2/core"
)

const (
	HookNameStartCoreService = "startCoreService"
)

var (
	Hooks = map[string]interface{}{}
)

func init() {
	rootCmd.AddCommand(coreCmd)
}

var coreCmd = &cobra.Command{
	Use:   "core",
	Short: "Run core program of Baetyl",
	Long:  `Baetyl runs the core program to sync with cloud and manage all applications.`,
	Run: func(_ *cobra.Command, _ []string) {
		// actual func is core.StartCoreService()
		Hooks[HookNameStartCoreService].(core.StartCoreServiceFunc)()
	},
}
