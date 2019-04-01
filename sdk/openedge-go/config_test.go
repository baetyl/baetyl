package openedge

import (
	"reflect"
	"testing"

	"github.com/baidu/openedge/utils"
	"github.com/stretchr/testify/assert"
)

func TestAppConfigValidate(t *testing.T) {
	cfg := `
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
	var cfgObj AppConfig
	err := utils.UnmarshalYAML([]byte(cfg), &cfgObj)
	assert.NoError(t, err)
}

func TestGetRemovedVolumes(t *testing.T) {
	type args struct {
		olds []VolumeInfo
		news []VolumeInfo
	}
	tests := []struct {
		name string
		args args
		want []VolumeInfo
	}{
		{
			name: "nil->nil",
			args: args{
				olds: nil,
				news: nil,
			},
			want: []VolumeInfo{},
		},
		{
			name: "nil->a",
			args: args{
				olds: nil,
				news: []VolumeInfo{
					VolumeInfo{
						Path: "a",
					},
				},
			},
			want: []VolumeInfo{},
		},
		{
			name: "a->nil",
			args: args{
				olds: []VolumeInfo{
					VolumeInfo{
						Path: "a",
					},
				},
				news: nil,
			},
			want: []VolumeInfo{
				VolumeInfo{
					Path: "a",
				},
			},
		},
		{
			name: "a->ab",
			args: args{
				olds: []VolumeInfo{
					VolumeInfo{
						Path: "a",
					},
				},
				news: []VolumeInfo{
					VolumeInfo{
						Path: "a",
					},
					VolumeInfo{
						Path: "b",
					},
				},
			},
			want: []VolumeInfo{},
		},
		{
			name: "a->b",
			args: args{
				olds: []VolumeInfo{
					VolumeInfo{
						Path: "a",
					},
				},
				news: []VolumeInfo{

					VolumeInfo{
						Path: "b",
					},
				},
			},
			want: []VolumeInfo{
				VolumeInfo{
					Path: "a",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetRemovedVolumes(tt.args.olds, tt.args.news); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRemovedVolumes() = %v, want %v", got, tt.want)
			}
		})
	}
}
