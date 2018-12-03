package config

import (
	"github.com/baidu/openedge/module"
	"github.com/baidu/openedge/trans"
	"github.com/baidu/openedge/utils"
	"github.com/juju/errors"
	validator "gopkg.in/validator.v2"
)

func init() {
	validator.SetValidationFunc("principals", principalsValidate)
	validator.SetValidationFunc("subscriptions", subscriptionsValidate)
}

// Config all config of edge
type Config struct {
	module.Config `yaml:",inline" json:",inline"`

	Listen      []string          `yaml:"listen" json:"listen"`
	Certificate trans.Certificate `yaml:"certificate" json:"certificate"`

	Principals    []Principal    `yaml:"principals" json:"principals" validate:"principals"`
	Subscriptions []Subscription `yaml:"subscriptions" json:"subscriptions" validate:"subscriptions"`

	Message  Message  `yaml:"message" json:"message"`
	Status   Status   `yaml:"status" json:"status"`
	Storage  Storage  `yaml:"storage" json:"storage"`
	Shutdown Shutdown `yaml:"shutdown" json:"shutdown"`
}

// NewConfig creates a new config
func NewConfig(in []byte) (*Config, error) {
	c := &Config{}
	err := utils.UnmarshalYAML(in, c)
	return c, errors.Trace(err)
}
