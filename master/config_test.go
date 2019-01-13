package master_test

import (
	"testing"
	"time"

	"github.com/baidu/openedge/master"
	"github.com/baidu/openedge/module/config"
	"github.com/baidu/openedge/module/utils"
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

	l := config.Logger{}
	defaults.Set(&l)

	type args struct {
		in  []byte
		out *master.Config
	}
	tests := []struct {
		name    string
		args    args
		want    *master.Config
		wantErr error
	}{
		{
			name: t.Name(),
			args: args{
				in:  []byte(confString),
				out: new(master.Config),
			},
			want: &master.Config{
				Modules: []config.Module{
					config.Module{
						Name:  "openedge_hub",
						Entry: "openedge_hub",
						Restart: config.Policy{
							Policy: "always",
							Backoff: config.Backoff{
								Min:    time.Second,
								Max:    time.Minute * 5,
								Factor: 2,
							},
						},
						Params:  []string{},
						Expose:  []string{},
						Volumes: []string{},
						Env:     map[string]string{},
						Logger:  l,
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
