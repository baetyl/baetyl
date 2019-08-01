package utils

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
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

func TestParseEnvs(t *testing.T) {
	const EnvHostIdKey = "OPENEDGE_HOST_ID"
	hostId := "test_host_id"
	err :=  os.Setenv(EnvHostIdKey, hostId)
	assert.NoError(t, err)
	confString := `
id: id
name: name
others: others
envs:
	{{.OPENEDGE_HOST_ID}}
modules:
  - name: m1
  - name: m2
    params:
      - arg1
      - arg2
`
	expectedStringTemp := `
id: id
name: name
others: others
envs:
	%s
modules:
  - name: m1
  - name: m2
    params:
      - arg1
      - arg2
`
	expectedString := fmt.Sprintf(expectedStringTemp, hostId)
	res, err := ParseEnvs([]byte(confString))
	resString := string(res)
	assert.Equal(t, expectedString, resString)
	assert.NoError(t, err)

	confString2 := `
id: id
name: name
others: others
envs:
	{{OPENEDGE_HOST_ID}}
modules:
  - name: m1
  - name: m2
    params:
      - arg1
      - arg2
`
	res2, err2 := ParseEnvs([]byte(confString2))
	assert.Equal(t, []byte(nil), res2)
	assert.Error(t, err2)
}
