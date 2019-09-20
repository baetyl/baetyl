package utils

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testEncodeStruct struct {
	Others  string             `yaml:"others" json:"others"`
	Modules []testEncodeModule `yaml:"modules" json:"modules" default:"[]"`
}

type testEncodeModule struct {
	Name   string   `yaml:"name" json:"name"  validate:"regexp=^(m1|m2)$"`
	Params []string `yaml:"params" json:"params" default:"[\"-c\", \"conf.yml\"]"`
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

	confString2 := `{
    "id": "id",
    "name": "name",
    "others": "others",
    "modules": [
        {
            "name": "k9"
        },
		{
            "name": "m2",
            "params": [
                "arg1",
                "arg2"
            ]
        }
    ]
}`
	err = UnmarshalYAML([]byte(confString2), &cfg2)
	assert.Error(t, err)
}

func TestParseEnv(t *testing.T) {
	const EnvHostIDKey = "BAETYL_HOST_ID"
	hostID := "test_host_id"
	err := os.Setenv(EnvHostIDKey, hostID)
	assert.NoError(t, err)
	confString := `
id: id
name: name
others: others
env: {{.BAETYL_HOST_ID}}
modules:
  - name: m1
  - name: m2
    params:
      - arg1
      - arg2
`
	expectedString := strings.Replace(confString, "{{.BAETYL_HOST_ID}}", hostID, 1)
	res, err := ParseEnv([]byte(confString))
	resString := string(res)
	assert.Equal(t, expectedString, resString)
	assert.NoError(t, err)

	// env not exist, env of parsed string would be empty
	confString2 := `
id: id
name: name
others: others
env: {{.BAETYL_NOT_EXIST}}
modules:
  - name: m1
  - name: m2
    params:
      - arg1
      - arg2
`
	_, err2 := ParseEnv([]byte(confString2))
	assert.Error(t, err2)

	// syntax error
	confString3 := `
id: id
name: name
others: others
env: {{BAETYL_HOST_ID}}
modules:
  - name: m1
  - name: m2
    params:
      - arg1
      - arg2
`
	res3, err3 := ParseEnv([]byte(confString3))
	assert.Equal(t, []byte(nil), res3)
	assert.Error(t, err3)
}

func TestUnmarshalJSON(t *testing.T) {
	confString := `{
    "id": "id",
    "name": "name",
    "others": "others",
    "modules": [
        {
            "name": "m1"
        },
		{
            "name": "m2",
            "params": [
                "arg1",
                "arg2"
            ]
        }
    ]
}`
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
	err := UnmarshalJSON([]byte(confString), &cfg2)
	assert.NoError(t, err)
	assert.Equal(t, cfg, cfg2)

	err = UnmarshalJSON([]byte("{"), &cfg2)
	assert.Error(t, err)

	confString2 := `{
    "id": "id",
    "name": "name",
    "others": "others",
    "modules": [
        {
            "name": "k9"
        },
		{
            "name": "m2",
            "params": [
                "arg1",
                "arg2"
            ]
        }
    ]
}`
	err = UnmarshalJSON([]byte(confString2), &cfg2)
	assert.Error(t, err)
}

func TestLoadYAML(t *testing.T) {
	dir, err := ioutil.TempDir("", "template")
	assert.NoError(t, err)
	fileName := "template_test"
	f, err := os.Create(filepath.Join(dir, fileName))
	defer f.Close()
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
	_, err = io.WriteString(f, confString)
	assert.NoError(t, err)

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
	err = LoadYAML(filepath.Join(dir, fileName), &cfg2)
	assert.NoError(t, err)
	assert.Equal(t, cfg, cfg2)

	fakeFileName := "fake"
	err = LoadYAML(filepath.Join(dir, fakeFileName), &cfg2)
	assert.Error(t, err)

	confString2 := `
id: id
name: name
others: others
env: {{BAETYL_HOST_ID}}
modules:
  - name: m1
  - name: m2
    params:
      - arg1
      - arg2
`
	fileName2 := "template_test2"
	f2, err := os.Create(filepath.Join(dir, fileName2))
	assert.NoError(t, err)
	_, err = io.WriteString(f2, confString2)
	err = LoadYAML(filepath.Join(dir, fileName2), &cfg2)
	assert.NoError(t, err)
	assert.Equal(t, cfg, cfg2)
}

func TestUnmarshalYAML(t *testing.T) {
	confString := "max: 2"
	l := Length{1}
	unmarshal := func(ls interface{}) error {
		err := UnmarshalYAML([]byte(confString), ls)
		if err != nil {
			return err
		}
		return nil
	}
	err := l.UnmarshalYAML(unmarshal)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), l.Max)

	confString2 := "max: \n2"
	unmarshal2 := func(ls interface{}) error {
		err := UnmarshalYAML([]byte(confString2), ls)
		if err != nil {
			return err
		}
		return nil
	}
	err = l.UnmarshalYAML(unmarshal2)
	assert.Error(t, err)
}
