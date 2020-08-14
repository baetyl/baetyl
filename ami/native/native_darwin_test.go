package native

import (
	"testing"

	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl/config"
	"github.com/stretchr/testify/assert"
)

func TestAmiNativeImpl(t *testing.T) {
	t.SkipNow()

	type args struct {
		ns      string
		app     v1.Application
		configs map[string]v1.Configuration
		secrets map[string]v1.Secret
	}
	tests := []struct {
		name          string
		args          args
		expectedStats []v1.AppStats
	}{
		{
			name: "appWithoutConfigAndSecret",
			args: args{
				ns: "baetyl-edge-system",
				app: v1.Application{
					Name:      "app1",
					Namespace: "xxx",
					Version:   "1111",
					Services: []v1.Service{
						{
							Name:    "svc1",
							Image:   "nginx",
							Replica: 1,
						},
					},
				},
			},
			expectedStats: []v1.AppStats{{
				AppInfo: v1.AppInfo{Name: "app1", Version: "1111"},
				Status:  "",
				Cause:   "",
				InstanceStats: map[string]v1.InstanceStats{
					"app1.1111.svc1.1": {
						Name:        "app1.1111.svc1.1",
						ServiceName: "svc1",
						Status:      "Unknown",
						Cause:       "\"launchctl\" failed with stderr: Could not find service \"app1.1111.svc1.1\" in domain for system\n",
					},
				},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl, err := newNativeImpl(config.AmiConfig{})
			assert.NoError(t, err)
			err = impl.ApplyApp(tt.args.ns, tt.args.app, tt.args.configs, tt.args.secrets)
			assert.NoError(t, err)

			stats, err := impl.StatsApps(tt.args.ns)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStats, stats)

			err = impl.DeleteApp(tt.args.ns, tt.args.app.Name)
			assert.NoError(t, err)
		})
	}
}
