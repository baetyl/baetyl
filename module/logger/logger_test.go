package logger

import (
	"reflect"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestNewEntry(t *testing.T) {
	type args struct {
		vs []string
	}
	tests := []struct {
		name string
		args args
		want logrus.Fields
	}{
		{
			name: "0",
			args: args{
				vs: []string{},
			},
			want: logrus.Fields{},
		},
		{
			name: "1",
			args: args{
				vs: []string{"k1"},
			},
			want: logrus.Fields{},
		},
		{
			name: "2",
			args: args{
				vs: []string{"k1", "v2"},
			},
			want: logrus.Fields{"k1": "v2"},
		},
		{
			name: "3",
			args: args{
				vs: []string{"k1", "v2", "k3"},
			},
			want: logrus.Fields{"k1": "v2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithFields(tt.args.vs...); !reflect.DeepEqual(got.entry.Data, tt.want) {
				t.Errorf("WithFields() = %v, want %v", got.entry.Data, tt.want)
			}
		})
	}
}
