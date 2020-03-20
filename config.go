package baetyl

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/baetyl/baetyl/logger"
)

// Config of baetyl
type Config struct {
	DataPath  string
	PidPath   string         `yaml:"pid" json:"pid"`
	Daemon    bool           `yaml:"daemon" json:"daemon" default:"true"`
	Logger    logger.LogInfo `yaml:"logger" json:"logger"`
	License   string         `yaml:"license" json:"license" default:"license.cert"`
	APIConfig struct {
		Network string        `yaml:"network" json:"network" default:"unix"`
		Address string        `yaml:"address" json:"address"`
		Timeout time.Duration `yaml:"timeout" json:"timeout" default:"5m"`
	} `yaml:"api" json:"api"`
	LegacyAPIConfig struct {
		Network string        `yaml:"network" json:"network" default:"unix"`
		Address string        `yaml:"address" json:"address"`
		Timeout time.Duration `yaml:"timeout" json:"timeout" default:"5m"`
	} `yaml:"server" json:"server"`
	Database struct {
		Driver string `yaml:"driver" json:"driver"`
	} `yaml:"database" json:"database" default:"{\"driver\":\"sqlite3\"}"`
	Grace  time.Duration `yaml:"grace" json:"grace" default:"30s"`
	Docker struct {
		APIVersion string `yaml:"api_version" json:"api_version" default:"1.38"`
	} `yaml:"docker" json:"docker"`
	Manage struct {
		Address  string `yaml:"url" json:"url" default:"https://iotedge.gz.baidubce.com/v3/edge/info"`
		ClientID string `yaml:"clientid" json:"clientid"`
		Report   struct {
			Interval time.Duration `yaml:"interval" json:"interval" default:"20s"`
			Timeout  time.Duration `yaml:"timeout" json:"timeout" default:"1m"`
		} `yaml:"report" json:"report"`
		Desire struct {
			Address  string        `yaml:"address" json:"address"`
			Username string        `yaml:"username" json:"username"`
			Interval time.Duration `yaml:"interval" json:"interval" default:"20s"`
			Timeout  time.Duration `yaml:"timeout" json:"timeout" default:"1m"`
		} `yaml:"desire" json:"desire"`
	} `yaml:"manage" json:"manage"`
	// cache config file path
	File string
}

func compilePath(orig, prefix, defval string) string {
	if orig == "" {
		return filepath.Join(prefix, defval)
	} else if !filepath.IsAbs(orig) {
		return filepath.Join(prefix, orig)
	} else {
		return orig
	}
}

func checkDirOpen(filename string, perm os.FileMode) error {
	return os.MkdirAll(filepath.Dir(filename), perm)
}

func checkFileOpen(filename string, flag int, perm os.FileMode) error {
	err := checkDirOpen(filename, 0755)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(filename, flag, perm)
	if err == nil {
		f.Close()
	}
	return err
}

// Validate configuration
func (c *Config) validate(prefix string) error {
	c.PidPath = compilePath(c.PidPath, prefix, DefaultPidPath)
	if err := checkDirOpen(c.PidPath, 0755); err != nil {
		return err
	}
	c.Logger.Path = compilePath(c.Logger.Path, prefix, DefaultLoggerPath)
	if err := checkDirOpen(c.Logger.Path, 0755); err != nil {
		return err
	}
	c.DataPath = compilePath(c.DataPath, prefix, DefaultDataPath)
	if err := os.MkdirAll(c.DataPath, 0750); err != nil {
		return err
	}
	if c.APIConfig.Network != "unix" || c.LegacyAPIConfig.Network != "unix" {
		return errors.New("this version only support unix socket")
	}
	c.APIConfig.Address = compilePath(c.APIConfig.Address, prefix, DefaultAPIAddress)
	if err := checkDirOpen(c.APIConfig.Address, 0755); err != nil {
		return err
	}
	c.LegacyAPIConfig.Address = compilePath(c.LegacyAPIConfig.Address, prefix, fmt.Sprintf("%s-legacy", DefaultAPIAddress))
	if err := checkDirOpen(c.LegacyAPIConfig.Address, 0755); err != nil {
		return err
	}
	return nil
}
