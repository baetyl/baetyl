package cmd

import (
	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/spf13/cobra"

	"github.com/baetyl/baetyl/v2/config"
	"github.com/baetyl/baetyl/v2/core"
	"github.com/baetyl/baetyl/v2/plugin"
)

func init() {
	rootCmd.AddCommand(coreCmd)
}

var coreCmd = &cobra.Command{
	Use:   "core",
	Short: "Run core program of Baetyl",
	Long:  `Baetyl runs the core program to sync with cloud and manage all applications.`,
	Run: func(_ *cobra.Command, _ []string) {
		startCoreService()
	},
}

func startCoreService() {
	context.Run(func(ctx context.Context) error {
		var cfg config.Config
		err := ctx.LoadCustomConfig(&cfg)
		if err != nil {
			return errors.Trace(err)
		}
		plugin.ConfFile = ctx.ConfFile()

		c, err := core.NewCore(cfg)
		if err != nil {
			return errors.Trace(err)
		}
		defer c.Close()

		ctx.Wait()
		return nil
	})
}
