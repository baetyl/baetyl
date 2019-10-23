package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_FormatPlatformInfo(t *testing.T) {
	tests := []struct {
		name     string
		hostinfo HostInfo
		result   string
	}{
		{
			name: "os empty",
			hostinfo: HostInfo{
				OS: "",
			},
			result: "unknown",
		},
		{
			name: "os not arm",
			hostinfo: HostInfo{
				OS:           "darwin",
				Architecture: "amd64",
			},
			result: "darwin/amd64",
		},
		{
			name: "os arm",
			hostinfo: HostInfo{
				OS:           "linux",
				Architecture: "arm",
				Variant:      "v7",
			},
			result: "linux/arm/v7",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := tt.hostinfo.FormatPlatformInfo()
			assert.Equal(t, res, tt.result)
		})
	}
}

func Test_parseGPUInfo(t *testing.T) {
	type args struct {
		in string
	}
	tests := []struct {
		name     string
		args     args
		wantGpus []PerGPUInfo
	}{
		{
			name: "normal",
			args: args{
				in: `
				0, TITAN X (Pascal), 12189, 12187, 0, 0
				1, TITAN X (Pascal), 12189, 12187, 12.3, 1`,
			},
			wantGpus: []PerGPUInfo{
				PerGPUInfo{
					Index:          "0",
					Model:          "TITAN X (Pascal)",
					MemTotal:       12189,
					MemFree:        12187,
					MemUsedPercent: 0.0,
					GPUUsedPercent: 0.0,
				},
				PerGPUInfo{
					Index:          "1",
					Model:          "TITAN X (Pascal)",
					MemTotal:       12189,
					MemFree:        12187,
					MemUsedPercent: 12.3,
					GPUUsedPercent: 1.0,
				},
			},
		},
		{
			name: "wrong",
			args: args{
				in: `
				0, TITAN X (Pascal), 12189, 12187, 0
				1, TITAN X (Pascal), 12189, 12187, 12.3, 1`,
			},
			wantGpus: []PerGPUInfo{
				PerGPUInfo{
					Index:          "1",
					Model:          "TITAN X (Pascal)",
					MemTotal:       12189,
					MemFree:        12187,
					MemUsedPercent: 12.3,
					GPUUsedPercent: 1.0,
				},
			},
		},
		{
			name: "exception",
			args: args{
				in: `
				0, TITAN X (Pascal), 12189, 12187, xxx, 0`,
			},
			wantGpus: []PerGPUInfo{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotGpus := parseGPUInfo(tt.args.in)
			if !reflect.DeepEqual(gotGpus, tt.wantGpus) {
				t.Errorf("parseGPUInfo() = %v, want %v", gotGpus, tt.wantGpus)
			}
		})
	}
}

func TestGetNetInfo(t *testing.T) {
	tests := []struct {
		name    string
		wantErr string
	}{
		{
			name: "local",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetNetInfo()
			assert.Equal(t, got.Error, tt.wantErr)
			data, _ := json.Marshal(got)
			fmt.Println(string(data))
		})
	}
}

func TestGetCPUInfo(t *testing.T) {
	tests := []struct {
		name    string
		wantErr string
	}{
		{
			name: "local1",
		},
		{
			name: "local2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetCPUInfo()
			assert.Equal(t, got.Error, tt.wantErr)
			data, _ := json.Marshal(got)
			fmt.Println(string(data))
		})
	}
}
