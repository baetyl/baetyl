package utils_test

import (
	"testing"

	"github.com/baidu/openedge/utils"
	"github.com/stretchr/testify/assert"
)

type confModule struct {
	Name   string   `yaml:"name"`
	Params []string `yaml:"params" default:"[\"-c\", \"conf.yml\"]"`
}

type confStruct struct {
	Others  string       `yaml:"others"`
	Modules []confModule `yaml:"modules" default:"[]"`
}

func TestUnmarshal(t *testing.T) {
	confString := `
id: id
name: name
others: others
modules:
  - name: m1
  - name: m2
    params:
      - arg1
      - arg2
`
	type args struct {
		in  []byte
		out *confStruct
	}
	tests := []struct {
		name    string
		args    args
		want    *confStruct
		wantErr error
	}{
		{
			name: "defaults-struct-slice",
			args: args{
				in:  []byte(confString),
				out: new(confStruct),
			},
			want: &confStruct{
				Others: "others",
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
			err := utils.UnmarshalYAML(tt.args.in, tt.args.out)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, tt.args.out)
		})
	}
}
