package main

import (
	"testing"
	"time"

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
}
