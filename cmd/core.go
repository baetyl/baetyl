package cmd

import (
	"os"

	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/spf13/cobra"

	"github.com/baetyl/baetyl/config"
	"github.com/baetyl/baetyl/core"
	"github.com/baetyl/baetyl/plugin"
)

const (
	macroBaetylCoreStorePath      = "BAETYL_CORE_STORE_PATH"
	macroBaetylObjectDownloadPath = "BAETYL_OBJECT_DOWNLOAD_PATH"
	macroBaetylHostRootPath       = "BAETYL_HOST_ROOT_PATH"
	macroBaetylNativeAppRunPath   = "BAETYL_NATIVE_APP_RUN_PATH"
)

var (
	envMap = map[string]string{
		macroBaetylCoreStorePath:      "/var/lib/baetyl/store",
		macroBaetylObjectDownloadPath: "/var/lib/baetyl/object",
		macroBaetylHostRootPath:       "/var/lib/baetyl/host",
		macroBaetylNativeAppRunPath:   "/var/lib/baetyl/run",
	}
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
		err := LoadConfig(ctx, &cfg)
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

func LoadConfig(ctx context.Context, cfg interface{}) error {
	for k, v := range envMap {
		if os.Getenv(k) == "" {
			err := os.Setenv(k, v)
			if err != nil {
				return errors.Trace(err)
			}
		}
	}
	return ctx.LoadCustomConfig(&cfg)
}
