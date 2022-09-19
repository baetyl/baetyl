package native

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/stretchr/testify/assert"

	"github.com/baetyl/baetyl/v2/ami"
	"github.com/baetyl/baetyl/v2/config"
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

func TestUpdateEnv(t *testing.T) {
	envs := []v1.Environment{
		{
			Name:  "a",
			Value: "va",
		},
		{
			Name:  "b",
			Value: "vb",
		},
	}
	envs = setEnv(envs, "a", "vaa")
	envs = setEnv(envs, "c", "vc")

	var f, ff int
	for _, v := range envs {
		if v.Name == "c" {
			f++
			assert.Equal(t, v.Value, "vc")
		}
		if v.Name == "a" {
			ff++
			assert.Equal(t, v.Value, "vaa")
		}
	}
	assert.Equal(t, f, 1)
	assert.Equal(t, ff, 1)
}

func TestNativeRemoteCommand(t *testing.T) {
	cfg := config.AmiConfig{}
	cfg.Native.PortsRange.Start, cfg.Native.PortsRange.End = 8000, 9000
	impl, err := newNativeImpl(cfg)
	assert.NoError(t, err)

	option := &ami.DebugOptions{}
	option.NativeDebugOptions.IP = "localhost"
	option.NativeDebugOptions.Port = "0"
	option.NativeDebugOptions.Username = "root"
	option.NativeDebugOptions.Password = "root"
	err = impl.RemoteCommand(option, ami.Pipe{})
	assert.NotEqual(t, err, nil)
}

func TestRPCApp(t *testing.T) {
	cfg := config.AmiConfig{}
	cfg.Native.PortsRange.Start, cfg.Native.PortsRange.End = 8000, 9000
	impl, err := newNativeImpl(cfg)
	assert.NoError(t, err)

	// req rpc fail
	req := &v1.RPCRequest{
		App:    "app",
		Method: "unknown",
		System: true,
		Params: "",
		Header: map[string]string{},
		Body:   "",
	}
	_, err = impl.RPCApp("", req)
	assert.NotNil(t, err)

	// req2 rpc get success
	req2 := &v1.RPCRequest{
		App:    "app",
		Method: "get",
		System: false,
		Params: "",
		Header: map[string]string{},
		Body:   "",
	}

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	port := strings.Split(s.URL, ":")
	req2.Params = ":" + port[2]
	_, err = impl.RPCApp(s.URL, req2)
	assert.NoError(t, err)
	s.Close()
}
