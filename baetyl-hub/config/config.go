package config

import (
	"time"

	"github.com/baetyl/baetyl/utils"
	validator "gopkg.in/validator.v2"
)

func init() {
	validator.SetValidationFunc("principals", principalsValidate)
	validator.SetValidationFunc("subscriptions", subscriptionsValidate)
}

// Config all config of edge
type Config struct {
	Listen      []string          `yaml:"listen" json:"listen"`
	Certificate utils.Certificate `yaml:"certificate" json:"certificate"`

	Principals    []Principal    `yaml:"principals" json:"principals" validate:"principals"`
	Subscriptions []Subscription `yaml:"subscriptions" json:"subscriptions" validate:"subscriptions"`

	Message Message `yaml:"message" json:"message"`
	Storage struct {
		Dir string `yaml:"dir" json:"dir" default:"var/db/baetyl/data"`
	} `yaml:"storage" json:"storage"`
	Shutdown struct {
		Timeout time.Duration `yaml:"timeout" json:"timeout" default:"10m"`
	} `yaml:"shutdown" json:"shutdown"`
	Metrics struct {
		Report struct {
			Interval time.Duration `yaml:"interval" json:"interval" default:"1m"`
		} `yaml:"report" json:"report"`
	} `yaml:"metrics" json:"metrics"`
}

// New config
func New(in []byte) (*Config, error) {
	c := &Config{}
	err := utils.UnmarshalYAML(in, c)
	return c, err
}
