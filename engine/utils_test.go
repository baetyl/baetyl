package engine

import (
	specv1 "github.com/baetyl/baetyl-go/spec/v1"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckService(t *testing.T) {
	tests := []struct {
		apps       map[string]specv1.Application
		stats      map[string]specv1.AppStats
		update     map[string]specv1.AppInfo
		expected   map[string]specv1.AppInfo
		statsNames []string
	}{
		{
			apps: map[string]specv1.Application{
				"app1": {
					Name: "app1",
					Services: []specv1.Service{{
						Name: "svc1",
					}},
				},
				"app2": {
					Name: "app2",
					Services: []specv1.Service{{
						Name: "svc1",
					}},
				},
			},
			stats: map[string]specv1.AppStats{},
			update: map[string]specv1.AppInfo{
				"app1": {
					Name:    "app1",
					Version: "v1",
				},
				"app2": {
					Name:    "app2",
					Version: "v1",
				},
			},
			expected: map[string]specv1.AppInfo{
				"app1": {
					Name:    "app1",
					Version: "v1",
				},
			},
			statsNames: []string{"app2"},
		},
		{
			apps: map[string]specv1.Application{
				"app1": {
					Name: "app1",
					Services: []specv1.Service{{
						Name: "svc1",
					}},
				},
				"app2": {
					Name: "app2",
					Services: []specv1.Service{{
						Name: "svc1",
					}},
				},
			},
			stats: map[string]specv1.AppStats{
				"app2": {},
			},
			update: map[string]specv1.AppInfo{
				"app1": {
					Name:    "app1",
					Version: "v1",
				},
				"app2": {
					Name:    "app2",
					Version: "v2",
				},
			},
			expected: map[string]specv1.AppInfo{
				"app2": {
					Name:    "app2",
					Version: "v2",
				},
			},
			statsNames: []string{"app1"},
		},
		{
			apps: map[string]specv1.Application{
				"app1": {
					Name: "app1",
					Services: []specv1.Service{{
						Name: "svc1",
						Ports: []specv1.ContainerPort{{
							HostPort: 1883,
						}},
					}},
				},
				"app2": {
					Name: "app2",
					Services: []specv1.Service{{
						Name: "svc2",
						Ports: []specv1.ContainerPort{{
							HostPort: 1883,
						}},
					}},
				},
			},
			stats: map[string]specv1.AppStats{"app2": {InstanceStats: map[string]specv1.InstanceStats{"svc2": {}}}},
			update: map[string]specv1.AppInfo{
				"app1": {
					Name:    "app1",
					Version: "v1",
				},
				"app2": {
					Name:    "app2",
					Version: "v2",
				},
			},
			expected: map[string]specv1.AppInfo{
				"app2": {
					Name:    "app2",
					Version: "v2",
				},
			},
			statsNames: []string{"app1"},
		},
		{
			apps: map[string]specv1.Application{
				"app1": {
					Name: "app1",
					Services: []specv1.Service{{
						Name: "svc1",
						Ports: []specv1.ContainerPort{{
							HostPort: 1883,
						}},
					}},
				},
				"app2": {
					Name: "app2",
					Services: []specv1.Service{{
						Name: "svc2",
						Ports: []specv1.ContainerPort{{
							HostPort: 1883,
						}},
					}},
				},
			},
			stats: map[string]specv1.AppStats{"app2": {InstanceStats: map[string]specv1.InstanceStats{"svc2": {}}}},
			update: map[string]specv1.AppInfo{
				"app1": {
					Name:    "app1",
					Version: "v1",
				},
				"app2": {
					Name:    "app2",
					Version: "v2",
				},
			},
			expected: map[string]specv1.AppInfo{
				"app2": {
					Name:    "app2",
					Version: "v2",
				},
			},
			statsNames: []string{"app1"},
		},
	}
	for _, tt := range tests {
		checkService(tt.apps, tt.stats, tt.update)
		checkPort(tt.apps, tt.stats, tt.update)
		assert.Equal(t, tt.update, tt.expected)
		for _, n := range tt.statsNames {
			assert.NotNil(t, tt.stats[n].InstanceStats)
		}
	}
}
