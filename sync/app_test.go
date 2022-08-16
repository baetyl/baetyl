package sync

import (
	"testing"

	"github.com/baetyl/baetyl-go/v2/context"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/stretchr/testify/assert"
)

func TestSync_PrepareApp(t *testing.T) {
	t.Setenv(context.KeyNodeName, "node01")

	app := specv1.Application{
		Name:    "app1",
		Version: "v1",
		Services: []specv1.Service{
			{
				Name: "s0",
			},
		},
		Volumes: []specv1.Volume{
			{
				Name: "v-1",
				VolumeSource: specv1.VolumeSource{
					HostPath: &specv1.HostPathVolumeSource{
						Path: "/var/lib/xxx",
					},
				},
			},
			{
				Name: "v-2",
				VolumeSource: specv1.VolumeSource{
					HostPath: &specv1.HostPathVolumeSource{
						Path: "var/lib/xxx",
					},
				},
			},
			{
				Name: "v-3",
				VolumeSource: specv1.VolumeSource{
					Config: &specv1.ObjectReference{
						Name:    "cfg-1",
						Version: "123",
					},
				},
			},
			{
				Name: "v-4",
				VolumeSource: specv1.VolumeSource{
					Config: &specv1.ObjectReference{
						Name:    "cfg-2",
						Version: "124",
					},
				},
			},
		},
	}

	configs := map[string]specv1.Configuration{}
	configs["cfg-1"] = specv1.Configuration{
		Name:    "cfg-1",
		Version: "123",
		Data: map[string]string{
			"file": "data",
		},
	}
	configs["cfg-2"] = specv1.Configuration{
		Name:    "cfg-2",
		Version: "124",
		Data: map[string]string{
			"_object_file": "{}",
		},
	}

	expApp := specv1.Application{
		Name:    "app1",
		Version: "v1",
		Services: []specv1.Service{
			{
				Name: "s0",
				Env: []specv1.Environment{
					{
						Name:  context.KeyAppName,
						Value: app.Name,
					},
					{
						Name:  context.KeySvcName,
						Value: "s0",
					},
					{
						Name:  context.KeyAppVersion,
						Value: app.Version,
					},
					{
						Name:  context.KeyNodeName,
						Value: "node01",
					},
					{
						Name:  context.KeyRunMode,
						Value: context.RunMode(),
					},
				},
			},
		},
		Volumes: []specv1.Volume{
			{
				Name: "v-1",
				VolumeSource: specv1.VolumeSource{
					HostPath: &specv1.HostPathVolumeSource{
						Path: "/var/lib/xxx",
					},
				},
			},
			{
				Name: "v-2",
				VolumeSource: specv1.VolumeSource{
					HostPath: &specv1.HostPathVolumeSource{
						Path: "var/lib/baetyl/host/var/lib/xxx",
					},
				},
			},
			{
				Name: "v-3",
				VolumeSource: specv1.VolumeSource{
					Config: &specv1.ObjectReference{
						Name:    "cfg-1",
						Version: "123",
					},
				},
			},
			{
				Name: "v-4",
				VolumeSource: specv1.VolumeSource{
					HostPath: &specv1.HostPathVolumeSource{
						Path: "var/lib/baetyl/object/cfg-2/124",
					},
				},
			},
		},
	}

	err := PrepareApp("var/lib/baetyl/host", "var/lib/baetyl/object", &app, configs)
	assert.NoError(t, err)
	assert.EqualValues(t, expApp, app)
}
