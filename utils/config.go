package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"github.com/docker/go-units"
	"gopkg.in/validator.v2"
	"gopkg.in/yaml.v2"
)

// LoadYAML config into out interface, with defaults and validates
func LoadYAML(path string, out interface{}) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	res, err := ParseEnv(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config parse error: %s", err.Error())
		res = data
	}
	return UnmarshalYAML(res, out)
}

// ParseEnv pasre env
func ParseEnv(data []byte) ([]byte, error) {
	text := string(data)
	envs := os.Environ()
	envMap := make(map[string]string)
	for _, s := range envs {
		t := strings.Split(s, "=")
		envMap[t[0]] = t[1]
	}
	tmpl, err := template.New("template").Option("missingkey=error").Parse(text)
	if err != nil {
		return nil, err
	}
	buffer := bytes.NewBuffer(nil)
	err = tmpl.Execute(buffer, envMap)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// UnmarshalYAML unmarshals, defaults and validates
func UnmarshalYAML(in []byte, out interface{}) error {
	err := yaml.Unmarshal(in, out)
	if err != nil {
		return err
	}
	err = SetDefaults(out)
	if err != nil {
		return err
	}
	err = validator.Validate(out)
	if err != nil {
		return err
	}
	return nil
}

// UnmarshalJSON unmarshals, defaults and validates
func UnmarshalJSON(in []byte, out interface{}) error {
	err := json.Unmarshal(in, out)
	if err != nil {
		return err
	}
	err = SetDefaults(out)
	if err != nil {
		return err
	}
	err = validator.Validate(out)
	if err != nil {
		return err
	}
	return nil
}

// Length length
type Length struct {
	Max int64 `yaml:"max" json:"max"`
}

// UnmarshalYAML customizes unmarshal
func (l *Length) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var ls length
	err := unmarshal(&ls)
	if err != nil {
		return err
	}
	if ls.Max != "" {
		l.Max, err = units.RAMInBytes(ls.Max)
		if err != nil {
			return err
		}
	}
	return nil
}

type length struct {
	Max string `yaml:"max" json:"max"`
}
