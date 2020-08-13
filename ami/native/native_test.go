package native

import (
	"github.com/baetyl/baetyl/config"
	"github.com/golangplus/testing/assert"
	"testing"

	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
)

func IgnoreTestAmiNativeImpl(t *testing.T) {
	type args struct {
		ns      string
		app     v1.Application
		configs map[string]v1.Configuration
		secrets map[string]v1.Secret
	}
	tests := []struct {
		name string
		args args
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl, err := newNativeImpl(config.AmiConfig{})
			assert.NoError(t, err)
			err = impl.ApplyApp(tt.args.ns, tt.args.app, tt.args.configs, tt.args.secrets)
			assert.NoError(t, err)

			err = impl.DeleteApp(tt.args.ns, tt.args.app.Name)
			assert.NoError(t, err)
		})
	}
}
