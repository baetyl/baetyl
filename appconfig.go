package baetyl

import (
	"path/filepath"

	schema0 "github.com/baetyl/baetyl/schema/v0"
	schema3 "github.com/baetyl/baetyl/schema/v3"
	"github.com/baetyl/baetyl/utils"
)

const appConfigFile = "application.yml"

func (rt *runtime) loadAppConfig(rootdir string) (*schema3.ComposeAppConfig, error) {
	cfg := new(schema3.ComposeAppConfig)
	configFile := filepath.Join(rootdir, appConfigFile)
	err := utils.LoadYAML(configFile, cfg)
	if err != nil {
		var c schema0.AppConfig
		err = utils.LoadYAML(configFile, &c)
		if err != nil {
			return cfg, err
		}
		cfg.FromAppConfig(&c)
	}
	return cfg, nil
}
