package openedge

import (
	"reflect"
	"testing"
)

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
