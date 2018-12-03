package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type confModule struct {
	Name   string   `yaml:"name"`
	Params []string `yaml:"params" default:"[\"-c\", \"conf.yml\"]"`
}

type confStruct struct {
	Others  string        `yaml:"others"`
	Timeout time.Duration `yaml:"timeout" default:"1m"`
	Modules []confModule  `yaml:"modules" default:"[]"`
}

func TestSetDefaults(t *testing.T) {
	tests := []struct {
		name    string
		args    *confStruct
		want    *confStruct
		wantErr bool
	}{
		{
			name: "defaults-struct-slice",
			args: &confStruct{
				Others: "others",
				Modules: []confModule{
					confModule{
						Name: "m1",
					},
					confModule{
						Name:   "m2",
						Params: []string{"arg1", "arg2"},
					},
				},
			},
			want: &confStruct{
				Others:  "others",
				Timeout: time.Minute,
				Modules: []confModule{
					confModule{
						Name:   "m1",
						Params: []string{"-c", "conf.yml"},
					},
					confModule{
						Name:   "m2",
						Params: []string{"arg1", "arg2"},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SetDefaults(tt.args); (err != nil) != tt.wantErr {
				t.Errorf("SetDefaults() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want, tt.args)
		})
	}
}
