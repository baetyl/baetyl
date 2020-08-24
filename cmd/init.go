package cmd

import (
	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	_ "github.com/baetyl/baetyl/ami"
	"github.com/baetyl/baetyl/config"
	"github.com/baetyl/baetyl/initz"
	"github.com/baetyl/baetyl/plugin"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Run init program of Baetyl",
	Long:  `Baetyl runs the init program to sync with cloud and start core service.`,
	Run: func(_ *cobra.Command, _ []string) {
		startInitService()
	},
}

func startInitService() {
	context.Run(func(ctx context.Context) error {
		var cfg config.Config
		err := ctx.LoadCustomConfig(&cfg)
		if err != nil {
			return errors.Trace(err)
		}
		plugin.ConfFile = ctx.ConfFile()

		init, err := initz.NewInitialize(cfg)
		if err != nil {
			return errors.Trace(err)
		}
		defer init.Close()

		ctx.Wait()
		return nil
	})
}
