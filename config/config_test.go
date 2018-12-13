package config_test

import (
	"testing"
	"time"

	"github.com/baidu/openedge/config"
	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/module"
	"github.com/baidu/openedge/utils"
	"github.com/creasty/defaults"
	"github.com/stretchr/testify/assert"
)

func TestUnmarshalYAML(t *testing.T) {
	confString := `
mode: docker
modules:
    - name: 'openedge_hub'
      entry: 'openedge_hub'
logger:
    console: true
    level: debug
`

	l := logger.Config{}
	defaults.Set(&l)

	type args struct {
		in  []byte
		out *config.Master
	}
	tests := []struct {
		name    string
		args    args
		want    *config.Master
		wantErr error
	}{
		{
			name: t.Name(),
			args: args{
				in:  []byte(confString),
				out: new(config.Master),
			},
			want: &config.Master{
				Modules: []config.Module{
					config.Module{
						Config: module.Config{
							Name:   "openedge_hub",
							Logger: l,
						},
						Entry: "openedge_hub",
						Restart: module.Policy{
							Policy: "always",
							Backoff: module.Backoff{
								Min:    time.Second,
								Max:    time.Minute * 5,
								Factor: 2,
							},
						},
						Params: []string{},
						Expose: []string{},
						Env:    map[string]string{},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := utils.UnmarshalYAML(tt.args.in, tt.args.out)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want.Modules[0], tt.args.out.Modules[0])
		})
	}
}
