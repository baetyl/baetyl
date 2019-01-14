package config

import (
	openedge "github.com/baidu/openedge/api/go"
	"github.com/baidu/openedge/utils"
	validator "gopkg.in/validator.v2"
)

func init() {
	validator.SetValidationFunc("principals", principalsValidate)
	validator.SetValidationFunc("subscriptions", subscriptionsValidate)
}

// Config all config of edge
type Config struct {
	Name   string           `yaml:"name" json:"name" validate:"nonzero"`
	Logger openedge.LogInfo `yaml:"logger" json:"logger"`

	Listen      []string          `yaml:"listen" json:"listen"`
	Certificate utils.Certificate `yaml:"certificate" json:"certificate"`

	Principals    []Principal    `yaml:"principals" json:"principals" validate:"principals"`
	Subscriptions []Subscription `yaml:"subscriptions" json:"subscriptions" validate:"subscriptions"`

	Message  Message  `yaml:"message" json:"message"`
	Status   Status   `yaml:"status" json:"status"`
	Storage  Storage  `yaml:"storage" json:"storage"`
	Shutdown Shutdown `yaml:"shutdown" json:"shutdown"`
}

// New config
func New(in []byte) (*Config, error) {
	c := &Config{}
	err := utils.UnmarshalYAML(in, c)
	return c, err
}
