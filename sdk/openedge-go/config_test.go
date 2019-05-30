package openedge

import (
	"reflect"
	"testing"

	"github.com/baidu/openedge/utils"
	"github.com/stretchr/testify/assert"
)

var cfgV1 = `
version: V5
services:
  - name: agent
    image: '  hub.baidubce.com/iottest/openedge-agent:latest'
    replica: 1
    mounts:
      - name: agent-conf-lruvr02ct-V1
        path: etc/openedge
        readonly: true
      - name: agent-cert-lruvr02ct-V1
        path: var/db/openedge/cert
        readonly: true
      - name: agent-volumes-lruvr02ct-V1
        path: var/db/openedge/volumes
      - name: agent-log-lruvr02ct-V1
        path: var/log/openedge
      - name: openedge-agent-bin-linux-amd64-V3
        path: lib/openedge/hub.baidubce.com/iottest/openedge-agent:latest
        readonly: true
volumes:
  - name: agent-conf-lruvr02ct-V1
    path: var/db/openedge/agent-conf-lruvr02ct/V1
  - name: agent-cert-lruvr02ct-V1
    path: var/db/openedge/agent-cert-lruvr02ct/V1
  - name: agent-volumes-lruvr02ct-V1
    path: var/db/openedge
  - name: agent-log-lruvr02ct-V1
    path: var/db/openedge/agent-log
  - name: openedge-agent-bin-linux-amd64-V3
    path: var/db/openedge/openedge-agent-bin-linux-amd64/V3
`

var cfgV2 = `
version: V5
services:
  - name: agent
    image: '  hub.baidubce.com/iottest/openedge-agent:latest'
    replica: 1
    mounts:
      - name: agent-conf-lruvr02ct-V2
        path: etc/openedge
        readonly: true
      - name: agent-cert-lruvr02ct-V1
        path: var/db/openedge/cert
        readonly: true
      - name: agent-volumes-lruvr02ct-V1
        path: var/db/openedge/volumes
      - name: agent-log-lruvr02ct-V1
        path: var/log/openedge
      - name: openedge-agent-bin-linux-amd64-V3
        path: lib/openedge/hub.baidubce.com/iottest/openedge-agent:latest
        readonly: true
  - name: hub
    image: 'hub.baidubce.com/iottest/openedge-hub:latest'
    replica: 1
    mounts:
      - name: hub-conf-V1
        path: etc/openedge
volumes:
  - name: agent-conf-lruvr02ct-V2
    path: var/db/openedge/agent-conf-lruvr02ct/V1
  - name: agent-cert-lruvr02ct-V1
    path: var/db/openedge/agent-cert-lruvr02ct/V1
  - name: agent-volumes-lruvr02ct-V1
    path: var/db/openedge
  - name: agent-log-lruvr02ct-V1
    path: var/db/openedge/agent-log
  - name: openedge-agent-bin-linux-amd64-V3
    path: var/db/openedge/openedge-agent-bin-linux-amd64/V3
  - name: hub-conf-V1
    path: var/db/openedge/openedge-hub-config/V1
  `

var cfgV3 = `
version: V5
`

var cfgV4 = `
version: V5
services:
  - name: agent
    image: '  hub.baidubce.com/iottest/openedge-agent:latest'
    replica: 1
    mounts:
      - name: agent-conf-lruvr02ct-V2
        path: etc/openedge
        readonly: true
      - name: agent-cert-lruvr02ct-V1
        path: var/db/openedge/cert
        readonly: true
      - name: agent-volumes-lruvr02ct-V1
        path: var/db/openedge/volumes
      - name: agent-log-lruvr02ct-V1
        path: var/log/openedge
      - name: openedge-agent-bin-linux-amd64-V3
        path: lib/openedge/hub.baidubce.com/iottest/openedge-agent:latest
        readonly: true
  - name: hub
    image: 'hub.baidubce.com/iottest/openedge-hub:latest'
    replica: 1
    mounts:
      - name: hub-conf-V1
        path: etc/openedge
volumes:
  - name: agent-conf-lruvr02ct-V2
    path: var/db/openedge/agent-conf-lruvr02ct/V2
  - name: agent-cert-lruvr02ct-V1
    path: var/db/openedge/agent-cert-lruvr02ct/V1
  - name: agent-volumes-lruvr02ct-V1
    path: var/db/openedge
  - name: agent-log-lruvr02ct-V1
    path: var/db/openedge/agent-log
  - name: openedge-agent-bin-linux-amd64-V3
    path: var/db/openedge/openedge-agent-bin-linux-amd64/V3
  - name: hub-conf-V1
    path: var/db/openedge/openedge-hub-config/V1
`

func TestAppConfigValidate(t *testing.T) {
	var cfgObj AppConfig
	err := utils.UnmarshalYAML([]byte(cfgV1), &cfgObj)
	assert.NoError(t, err)
}

func TestDiffServices(t *testing.T) {
	var V1 AppConfig
	err := utils.UnmarshalYAML([]byte(cfgV1), &V1)
	assert.NoError(t, err)

	var V2 AppConfig
	err = utils.UnmarshalYAML([]byte(cfgV2), &V2)
	assert.NoError(t, err)

	var V3 AppConfig
	err = utils.UnmarshalYAML([]byte(cfgV3), &V3)
	assert.NoError(t, err)

	var V4 AppConfig
	err = utils.UnmarshalYAML([]byte(cfgV4), &V4)
	assert.NoError(t, err)

	servicesV1 := V1.Services
	servicesV2 := V2.Services
	servicesV3 := V3.Services
	servicesV4 := V4.Services

	volumesV1 := V1.Volumes
	volumesV2 := V2.Volumes
	volumesV3 := V3.Volumes
	volumesV4 := V4.Volumes

	serviceMap := make(map[string]ServiceInfo)
	for _, service := range servicesV1 {
		serviceMap[service.Name] = service
	}

	removedWant := []ServiceInfo{
		serviceMap["agent"],
	}

	for _, service := range servicesV2 {
		serviceMap[service.Name] = service
	}

	updatedWant := []ServiceInfo{
		serviceMap["agent"],
		serviceMap["hub"],
	}

	updatedVolumes, _ := DiffVolumes(volumesV1, volumesV2)
	updated, removed := DiffServices(servicesV1, servicesV2, updatedVolumes)
	if !reflect.DeepEqual(removed, removedWant) {
		t.Errorf("DiffServices() = %v, want %v", removed, removedWant)
	}
	if !reflect.DeepEqual(updated, updatedWant) {
		t.Errorf("DiffServices() = %v, want %v", updated, updatedWant)
	}

	removedWant = []ServiceInfo{
		serviceMap["agent"],
		serviceMap["hub"],
	}

	for _, service := range servicesV1 {
		serviceMap[service.Name] = service
	}

	updatedWant = []ServiceInfo{
		serviceMap["agent"],
	}

	updatedVolumes, _ = DiffVolumes(volumesV2, volumesV1)
	updated, removed = DiffServices(servicesV2, servicesV1, updatedVolumes)
	if !reflect.DeepEqual(removed, removedWant) {
		t.Errorf("DiffServices () = %v, want %v", removed, removedWant)
	}
	if !reflect.DeepEqual(updated, updatedWant) {
		t.Errorf("DiffServices () = %v, want %v", updated, updatedWant)
	}

	for _, service := range servicesV2 {
		serviceMap[service.Name] = service
	}

	removedWant = []ServiceInfo{
		serviceMap["agent"],
		serviceMap["hub"],
	}

	updatedWant = []ServiceInfo{}

	updatedVolumes, _ = DiffVolumes(volumesV2, volumesV3)
	updated, removed = DiffServices(servicesV2, servicesV3, updatedVolumes)
	if !reflect.DeepEqual(removed, removedWant) {
		t.Errorf("DiffServices () = %v, want %v", removed, removedWant)
	}
	if !reflect.DeepEqual(updated, updatedWant) {
		t.Errorf("DiffServices () = %v, want %v", updated, updatedWant)
	}

	updatedWant = removedWant
	removedWant = []ServiceInfo{}

	updatedVolumes, _ = DiffVolumes(volumesV3, volumesV2)
	updated, removed = DiffServices(servicesV3, servicesV2, updatedVolumes)
	if !reflect.DeepEqual(removed, removedWant) {
		t.Errorf("DiffServices () = %v, want %v", removed, removedWant)
	}
	if !reflect.DeepEqual(updated, updatedWant) {
		t.Errorf("DiffServices () = %v, want %v", updated, updatedWant)
	}

	for _, service := range servicesV2 {
		serviceMap[service.Name] = service
	}

	removedWant = []ServiceInfo{
		serviceMap["agent"],
	}

	for _, service := range servicesV4 {
		serviceMap[service.Name] = service
	}

	updatedWant = []ServiceInfo{
		serviceMap["agent"],
	}

	updatedVolumes, _ = DiffVolumes(volumesV2, volumesV4)
	updated, removed = DiffServices(servicesV2, servicesV4, updatedVolumes)
	if !reflect.DeepEqual(removed, removedWant) {
		t.Errorf("DiffServices () = %v, want %v", removed, removedWant)
	}
	if !reflect.DeepEqual(updated, updatedWant) {
		t.Errorf("DiffServices () = %v, want %v", updated, updatedWant)
	}
}

func TestDiffVolumes(t *testing.T) {
	type args struct {
		olds []VolumeInfo
		news []VolumeInfo
	}
	tests := []struct {
		name    string
		args    args
		removed []VolumeInfo
		updated map[string]bool
	}{
		{
			name: "nil->nil",
			args: args{
				olds: nil,
				news: nil,
			},
			removed: []VolumeInfo{},
			updated: make(map[string]bool),
		},
		{
			name: "nil->a",
			args: args{
				olds: nil,
				news: []VolumeInfo{
					VolumeInfo{
						Name: "a",
						Path: "a",
					},
				},
			},
			removed: []VolumeInfo{},
			updated: map[string]bool{
				"a": true,
			},
		},
		{
			name: "a->nil",
			args: args{
				olds: []VolumeInfo{
					VolumeInfo{
						Name: "a",
						Path: "a",
					},
				},
				news: nil,
			},
			removed: []VolumeInfo{
				VolumeInfo{
					Name: "a",
					Path: "a",
				},
			},
			updated: make(map[string]bool),
		},
		{
			name: "a->ab",
			args: args{
				olds: []VolumeInfo{
					VolumeInfo{
						Name: "a",
						Path: "a",
					},
				},
				news: []VolumeInfo{
					VolumeInfo{
						Name: "a",
						Path: "a",
					},
					VolumeInfo{
						Name: "b",
						Path: "b",
					},
				},
			},
			removed: []VolumeInfo{},
			updated: map[string]bool{
				"b": true,
			},
		},
		{
			name: "a->b",
			args: args{
				olds: []VolumeInfo{
					VolumeInfo{
						Name: "a",
						Path: "a",
					},
				},
				news: []VolumeInfo{

					VolumeInfo{
						Name: "b",
						Path: "b",
					},
				},
			},
			removed: []VolumeInfo{
				VolumeInfo{
					Name: "a",
					Path: "a",
				},
			},
			updated: map[string]bool{
				"b": true,
			},
		},
		{
			name: "aa->ab",
			args: args{
				olds: []VolumeInfo{
					VolumeInfo{
						Name: "a",
						Path: "a",
					},
				},
				news: []VolumeInfo{
					VolumeInfo{
						Name: "a",
						Path: "b",
					},
				},
			},
			removed: []VolumeInfo{
				VolumeInfo{
					Name: "a",
					Path: "a",
				},
			},
			updated: map[string]bool{
				"a": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated, removed := DiffVolumes(tt.args.olds, tt.args.news)
			if !reflect.DeepEqual(removed, tt.removed) {
				t.Errorf("DiffVolumes() = %v, want %v", removed, tt.removed)
			}
			if !reflect.DeepEqual(updated, tt.updated) {
				t.Errorf("DiffVolumes() = %v, want %v", updated, tt.updated)
			}
		})
	}
}
