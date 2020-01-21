// +build !linux

package master

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"testing"
	"time"

	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/protocol/http"
	"github.com/baetyl/baetyl/sdk/baetyl-go"
	"github.com/baetyl/baetyl/utils"
	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	tests := []struct {
		name string
		args []byte
	}{
		{
			name: "nil",
			args: nil,
		},
		{
			name: "empty",
			args: []byte{},
		},
		{
			name: "empty2",
			args: []byte(""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg Config
			err := utils.UnmarshalYAML(tt.args, &cfg)
			assert.NoError(t, err)

			assert.Equal(t, "docker", cfg.Mode)

			if runtime.GOOS == "linux" {
				assert.Equal(t, "unix:///var/run/baetyl.sock", cfg.Server.Address)
			} else {
				assert.Equal(t, "tcp://127.0.0.1:50050", cfg.Server.Address)
			}
			assert.Equal(t, time.Duration(5*60*1000*1000000), cfg.Server.Timeout)

			assert.Equal(t, "var/log/baetyl/baetyl.log", cfg.Logger.Path)
			assert.Equal(t, "info", cfg.Logger.Level)
			assert.Equal(t, "text", cfg.Logger.Format)
			assert.Equal(t, 15, cfg.Logger.Age.Max)
			assert.Equal(t, 50, cfg.Logger.Size.Max)
			assert.Equal(t, 15, cfg.Logger.Backup.Max)

			assert.Equal(t, time.Duration(30*1000*1000000), cfg.Grace)
		})
	}
}

func TestConfig(t *testing.T) {
	dir, err := ioutil.TempDir("", t.Name())
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	conf := Config{
		Server: http.ServerInfo{
			Address: "baetyl://127.0.0.1:51150",
		},
	}

	err = conf.Validate()
	assert.Error(t, err)
	assert.Equal(t, "only support unix domian socket or tcp socket", err.Error())

	filepath := path.Join(dir, "sn")
	sn := "baetyl"
	f, err := os.Create(filepath)
	assert.NoError(t, err)

	n, err := io.WriteString(f, sn)
	assert.NoError(t, err)
	assert.Len(t, sn, n)
	f.Sync()
	f.Close()

	conf = Config{
		Server: http.ServerInfo{
			Address: "unix://127.0.0.1:51150",
		},
	}
	err = conf.Validate()
	assert.Error(t, err)
	assert.Equal(t, "unix domain socket only support on linux, please to use tcp socket", err.Error())

	conf = Config{
		Mode: "docker",
		Server: http.ServerInfo{
			Address: "tcp://127.0.0.1:51150",
		},
		SNFile: filepath,
	}
	err = conf.Validate()
	assert.NoError(t, err)
	assert.Equal(t, "tcp://host.docker.internal:51150", utils.GetEnv(baetyl.EnvKeyMasterAPIAddress))
	assert.Equal(t, sn, utils.GetEnv(baetyl.EnvKeyHostSN))
	assert.Equal(t, "v1", utils.GetEnv(baetyl.EnvKeyMasterAPIVersion))
	assert.Equal(t, runtime.GOOS, utils.GetEnv(baetyl.EnvKeyHostOS))
	assert.Equal(t, conf.Mode, utils.GetEnv(baetyl.EnvKeyServiceMode))

	conf = Config{
		Mode: "native",
		Server: http.ServerInfo{
			Address: "tcp://127.0.0.1:51150",
		},
		SNFile: filepath,
	}
	err = conf.Validate()
	assert.NoError(t, err)
	assert.Equal(t, conf.Server.Address, utils.GetEnv(baetyl.EnvKeyMasterAPIAddress))
	assert.Equal(t, sn, utils.GetEnv(baetyl.EnvKeyHostSN))
	assert.Equal(t, "v1", utils.GetEnv(baetyl.EnvKeyMasterAPIVersion))
	assert.Equal(t, runtime.GOOS, utils.GetEnv(baetyl.EnvKeyHostOS))
	assert.Equal(t, conf.Mode, utils.GetEnv(baetyl.EnvKeyServiceMode))
}

func TestOTALog(t *testing.T) {
	var cfg Config
	err := utils.UnmarshalYAML(nil, &cfg)
	assert.NoError(t, err)

	cfg.OTALog.Path = "testdata/ota.log"
	os.RemoveAll(cfg.OTALog.Path)
	defer os.RemoveAll(cfg.OTALog.Path)
	defer os.RemoveAll("testdata/ota.log.old")
	assert.False(t, utils.FileExists(cfg.OTALog.Path))
	logger.New(cfg.OTALog).WithField("step", "RECEIVE").WithField("trace", "xxxxxx").WithField("type", "APP").Infof("receive update event")
	assert.True(t, utils.FileExists(cfg.OTALog.Path))
	os.Rename(cfg.OTALog.Path, "testdata/ota.log.old")
	assert.False(t, utils.FileExists(cfg.OTALog.Path))
	logger.New(cfg.OTALog).WithField("step", "SUCCESS").WithField("trace", "xxxxxx").WithField("type", "APP").Infof("update application successfully")
	assert.True(t, utils.FileExists(cfg.OTALog.Path))
}
