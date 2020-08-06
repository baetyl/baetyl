package cmd

import (
	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl/config"
	"github.com/baetyl/baetyl/core"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(coreCmd)
}

var coreCmd = &cobra.Command{
	Use:   "core",
	Short: "Start core app of Baetyl",
	Long:  `Baetyl starts the core app to sync with cloud and manage all applications`,
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

		c, err := core.NewCore(cfg)
		if err != nil {
			return errors.Trace(err)
		}
		defer c.Close()

		ctx.Wait()
		return nil
	})
}
