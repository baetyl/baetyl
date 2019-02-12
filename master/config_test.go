package master

import (
	"reflect"
	"testing"
)

func TestDynamicConfig_diff(t *testing.T) {
	tests := []struct {
		name  string
		cur   *DynamicConfig
		pre   *DynamicConfig
		want  *dynamicConfigDiff
		want1 bool
	}{
		{
			name:  "empty",
			cur:   new(DynamicConfig),
			pre:   new(DynamicConfig),
			want:  nil,
			want1: true,
		},
		{
			name: "add",
			cur: &DynamicConfig{
				Version: "v1",
			},
			pre:   new(DynamicConfig),
			want:  nil,
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.cur.diff(tt.pre)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DynamicConfig.diff() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("DynamicConfig.diff() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
