package kube

import (
	"testing"

	"helm.sh/helm/v3/pkg/release"

	"github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/stretchr/testify/assert"
)

func TestSetChartValues(t *testing.T) {
	tests := []struct {
		name      string
		app       v1.Application
		cfgs      map[string]v1.Configuration
		expectDir string
		expectVal map[string]interface{}
		err       error
	}{
		{
			name: "no chart tar has been selected",
			app: v1.Application{
				Services: []v1.Service{
					{
						Image:   "test.tgz",
						Runtime: "",
					},
				},
				Volumes: []v1.Volume{
					{
						VolumeSource: v1.VolumeSource{
							HostPath: &v1.HostPathVolumeSource{
								Path: "/tmp",
							},
						},
					},
				},
			},
			cfgs:      map[string]v1.Configuration{},
			expectDir: "/tmp/test.tgz",
			expectVal: nil,
			err:       nil,
		},
		{
			name: "config not exist",
			app: v1.Application{
				Services: []v1.Service{
					{
						Image:   "test.tgz",
						Runtime: "values.yaml",
					},
				},
				Volumes: []v1.Volume{
					{
						VolumeSource: v1.VolumeSource{
							HostPath: &v1.HostPathVolumeSource{
								Path: "/tmp",
							},
						},
					},
					{
						VolumeSource: v1.VolumeSource{
							Config: &v1.ObjectReference{
								Name:    "config1",
								Version: "12345",
							},
						},
					},
				},
			},
			cfgs: map[string]v1.Configuration{
				"config1": {
					Data: map[string]string{
						"values.yaml": `
image:
  repository: nginx
  tag: 1.20
`,
					},
				},
			},
			expectDir: "/tmp/test.tgz",
			expectVal: map[string]interface{}{
				"image": map[string]interface{}{
					"repository": "nginx",
					"tag":        1.20,
				},
			},
			err: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, val, err := setChartValues(tt.app, tt.cfgs)
			assert.Equal(t, err, tt.err)
			assert.Equal(t, dir, tt.expectDir)
			assert.Equal(t, val, tt.expectVal)
		})
	}
}

func TestTransStatus(t *testing.T) {
	statuses := []struct {
		name     string
		status   release.Status
		expected v1.Status
	}{
		{
			name:     "status unknown",
			status:   release.StatusUnknown,
			expected: v1.Unknown,
		},
		{
			name:     "status deployed",
			status:   release.StatusDeployed,
			expected: v1.Running,
		},
		{
			name:     "status failed",
			status:   release.StatusFailed,
			expected: v1.Failed,
		},
		{
			name:     "status pending",
			status:   release.StatusPendingInstall,
			expected: v1.Pending,
		},
	}

	for _, tt := range statuses {
		t.Run(tt.name, func(t *testing.T) {
			res := transStatus(tt.status)
			assert.Equal(t, tt.expected, res, "should be equal")
		})
	}
}
