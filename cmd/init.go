package cmd

import (
	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	_ "github.com/baetyl/baetyl/ami"
	"github.com/baetyl/baetyl/config"
	"github.com/baetyl/baetyl/initz"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Start baetyl init app",
	Long:  `Baetyl starts the int app to sync with cloud and start core app`,
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

		init, err := initz.NewInitialize(cfg)
		if err != nil {
			return errors.Trace(err)
		}
		defer init.Close()

		ctx.Wait()
		return nil
	})
}
