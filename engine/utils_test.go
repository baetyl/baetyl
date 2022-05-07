package engine

import (
	"fmt"
	"testing"

	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/stretchr/testify/assert"
)

func TestCheckService(t *testing.T) {
	tests := []struct {
		infos      []specv1.AppInfo
		apps       map[string]specv1.Application
		stats      map[string]specv1.AppStats
		update     map[string]specv1.AppInfo
		expected   map[string]specv1.AppInfo
		statsNames []string
	}{
		{
			infos: []specv1.AppInfo{
				{Name: "app1", Version: "v1"},
				{Name: "app2", Version: "v1"},
			},
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
			infos: []specv1.AppInfo{
				{Name: "app1", Version: "v1"},
				{Name: "app2", Version: "v2"},
			},
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
			infos: []specv1.AppInfo{
				{Name: "app1", Version: "v1"},
				{Name: "app2", Version: "v2"},
			},
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
			infos: []specv1.AppInfo{
				{Name: "app1", Version: "v1"},
				{Name: "app2", Version: "v2"},
			},
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
			infos: []specv1.AppInfo{
				{Name: "app1", Version: "v1"},
			},
			apps: map[string]specv1.Application{
				"app1": {
					Name: "app1",
					Services: []specv1.Service{{
						Name: "svc1",
						Ports: []specv1.ContainerPort{{
							HostPort: 1883,
						}},
						Replica: 3,
					}},
				},
			},
			stats: map[string]specv1.AppStats{},
			update: map[string]specv1.AppInfo{
				"app1": {
					Name:    "app1",
					Version: "v1",
				},
			},
			expected:   map[string]specv1.AppInfo{},
			statsNames: []string{"app1"},
		},
	}
	for i, tt := range tests {
		checkService(tt.infos, tt.apps, tt.stats, tt.update)
		checkPort(tt.infos, tt.apps, tt.stats, tt.update)
		assert.Equal(t, tt.expected, tt.update, i)
		for _, n := range tt.statsNames {
			assert.NotNil(t, tt.stats[n].InstanceStats)
		}
	}
}

func TestCheckApps(t *testing.T) {
	tests := []struct {
		infos      []specv1.AppInfo
		apps       map[string]specv1.Application
		stats      map[string]specv1.AppStats
		update     map[string]specv1.AppInfo
		expected   map[string]specv1.AppInfo
		statsNames []string
	}{
		{
			infos: []specv1.AppInfo{
				{Name: "app1", Version: "v1"},
				{Name: "app2", Version: "v1"},
			},
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
				"app2": {
					Name:    "app2",
					Version: "v1",
				},
			},
			statsNames: []string{},
		},
		{
			infos: []specv1.AppInfo{
				{Name: "app1", Version: "v1"},
				{Name: "app2", Version: "v2"},
			},
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
				"app2": {InstanceStats: map[string]specv1.InstanceStats{"app2": {}}},
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
				"app1": {
					Name:    "app1",
					Version: "v1",
				},
				"app2": {
					Name:    "app2",
					Version: "v2",
				},
			},
			statsNames: []string{"app2"},
		},
		{
			infos: []specv1.AppInfo{
				{Name: "app1", Version: "v1"},
				{Name: "app2", Version: "v2"},
			},
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
			stats: map[string]specv1.AppStats{"app2": {InstanceStats: map[string]specv1.InstanceStats{"app2": {}}}},
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
			statsNames: []string{"app2"},
		},
		{
			infos: []specv1.AppInfo{
				{Name: "app1", Version: "v1"},
			},
			apps: map[string]specv1.Application{
				"app1": {
					Replica: 2,
					Name:    "app1",
					Services: []specv1.Service{{
						Name: "svc1",
						Ports: []specv1.ContainerPort{{
							HostPort: 1883,
						}},
					}},
				},
			},
			stats: map[string]specv1.AppStats{},
			update: map[string]specv1.AppInfo{
				"app1": {
					Name:    "app1",
					Version: "v1",
				},
			},
			expected:   map[string]specv1.AppInfo{},
			statsNames: []string{"app1"},
		},
		{
			infos: []specv1.AppInfo{
				{Name: "app1", Version: "v1"},
				{Name: "app2", Version: "v2"},
			},
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
			stats: map[string]specv1.AppStats{"app2": {InstanceStats: map[string]specv1.InstanceStats{"app2": {}}}},
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
			statsNames: []string{"app2"},
		},
		{
			infos: []specv1.AppInfo{
				{Name: "app1", Version: "v1"},
				{Name: "app2", Version: "v2"},
			},
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
							HostPort: 1884,
						}},
					}},
				},
			},
			stats: map[string]specv1.AppStats{"app2": {InstanceStats: map[string]specv1.InstanceStats{"app2": {}}}},
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
				"app1": {
					Name:    "app1",
					Version: "v1",
				},
				"app2": {
					Name:    "app2",
					Version: "v2",
				},
			},
			statsNames: []string{"app2"},
		},
	}
	for i, tt := range tests {
		fmt.Println(i)
		checkMultiAppPort(tt.infos, tt.apps, tt.stats, tt.update)
		assert.Equal(t, tt.expected, tt.update, i)
		for _, n := range tt.statsNames {
			assert.NotNil(t, tt.stats[n].InstanceStats)
		}
	}
}
