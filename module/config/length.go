package config

import (
	units "github.com/docker/go-units"
)

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
