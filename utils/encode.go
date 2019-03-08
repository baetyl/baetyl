package utils

import (
	"encoding/json"
	"io/ioutil"

	units "github.com/docker/go-units"
	validator "gopkg.in/validator.v2"
	yaml "gopkg.in/yaml.v2"
)

// LoadYAML config into out interface, with defaults and validates
func LoadYAML(path string, out interface{}) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return UnmarshalYAML(data, out)
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
