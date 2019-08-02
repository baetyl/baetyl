package master

import (
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/utils"
	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
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
				assert.Equal(t, "unix:///var/run/openedge.sock", cfg.Server.Address)
			} else {
				assert.Equal(t, "tcp://127.0.0.1:50050", cfg.Server.Address)
			}
			assert.Equal(t, time.Duration(5*60*1000*1000000), cfg.Server.Timeout)

			assert.Equal(t, "var/log/openedge/openedge.log", cfg.Logger.Path)
			assert.Equal(t, "info", cfg.Logger.Level)
			assert.Equal(t, "text", cfg.Logger.Format)
			assert.Equal(t, 15, cfg.Logger.Age.Max)
			assert.Equal(t, 50, cfg.Logger.Size.Max)
			assert.Equal(t, 15, cfg.Logger.Backup.Max)

			assert.Equal(t, time.Duration(30*1000*1000000), cfg.Grace)
		})
	}
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
