package utils

import (
	"encoding/json"

	"github.com/juju/errors"
	validator "gopkg.in/validator.v2"
	yaml "gopkg.in/yaml.v2"
)

// UnmarshalYAML unmarshals, defaults and validates
func UnmarshalYAML(in []byte, out interface{}) error {
	err := yaml.Unmarshal(in, out)
	if err != nil {
		return errors.Trace(err)
	}
	err = SetDefaults(out)
	if err != nil {
		return errors.Trace(err)
	}
	err = validator.Validate(out)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

// UnmarshalJSON unmarshals, defaults and validates
func UnmarshalJSON(in []byte, out interface{}) error {
	err := json.Unmarshal(in, out)
	if err != nil {
		return errors.Trace(err)
	}
	err = SetDefaults(out)
	if err != nil {
		return errors.Trace(err)
	}
	err = validator.Validate(out)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}
