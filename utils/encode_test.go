package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testEncodeStruct struct {
	Others  string             `yaml:"others"`
	Modules []testEncodeModule `yaml:"modules" default:"[]"`
}

type testEncodeModule struct {
	Name   string   `yaml:"name"`
	Params []string `yaml:"params" default:"[\"-c\", \"conf.yml\"]"`
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
	cfg := testEncodeStruct{
		Others: "others",
		Modules: []testEncodeModule{
			testEncodeModule{
				Name:   "m1",
				Params: []string{"-c", "conf.yml"},
			},
			testEncodeModule{
				Name:   "m2",
				Params: []string{"arg1", "arg2"},
			},
		},
	}
	var cfg2 testEncodeStruct
	err := UnmarshalYAML([]byte(confString), &cfg2)
	assert.NoError(t, err)
	assert.Equal(t, cfg, cfg2)

	err = UnmarshalYAML([]byte("-{}-"), &cfg2)
	assert.EqualError(t, err, "yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `-{}-` into utils.testEncodeStruct")
}
