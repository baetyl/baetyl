package config

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/baetyl/baetyl/baetyl-agent/common"
	"github.com/baetyl/baetyl/sdk/baetyl-go"
	"github.com/baetyl/baetyl/utils"
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg Config
			err := utils.UnmarshalYAML(tt.args, &cfg)
			assert.NoError(t, err)

			assert.Equal(t, "var/db/baetyl/volumes/ota.log", cfg.OTA.Logger.Path)
			assert.Equal(t, time.Duration(20*time.Second), cfg.Remote.Report.Interval)
			assert.Equal(t, time.Duration(5*time.Minute), cfg.OTA.Timeout)
		})
	}

	var res Resource

	expectedApp := Application{
		AppConfig: baetyl.ComposeAppConfig{
			Name:       "app",
			AppVersion: "v1",
		},
	}
	expectedRes := Resource{
		BaseResource: BaseResource{
			Type:    common.Application,
			Name:    "app",
			Version: "v1",
		},
		Value: expectedApp,
	}
	b, _ := json.Marshal(expectedRes)
	expectedRes.Data = b
	err := json.Unmarshal(b, &res)
	assert.NoError(t, err)
	assert.Equal(t, res.Data, b)
	app := res.GetApplication()
	assert.Equal(t, *app, expectedApp)

	expectedConfig := ModuleConfig{
		Name: "config",
		Data: map[string]string{
			"service.yml": "config",
		},
	}
	expectedRes = Resource{
		BaseResource: BaseResource{
			Type:    common.Config,
			Name:    "config",
			Version: "v1",
		},
		Value: expectedConfig,
	}
	b, _ = json.Marshal(expectedRes)
	expectedRes.Data = b
	err = json.Unmarshal(b, &res)
	assert.NoError(t, err)
	assert.Equal(t, res.Data, b)
	config := res.GetConfig()
	assert.Equal(t, *config, expectedConfig)
}
