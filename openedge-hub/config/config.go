package config

import (
	"time"

	"github.com/baidu/openedge/utils"
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
	Status  struct {
		Logging struct {
			Enable   bool          `yaml:"enable" json:"enable"`
			Interval time.Duration `yaml:"interval" json:"interval" default:"1m"`
		} `yaml:"logging" json:"logging"`
	} `yaml:"status" json:"status"`
	Storage struct {
		Dir string `yaml:"dir" json:"dir" default:"var/db/openedge"`
	} `yaml:"storage" json:"storage"`
	Shutdown struct {
		Timeout time.Duration `yaml:"timeout" json:"timeout" default:"10m"`
	} `yaml:"shutdown" json:"shutdown"`
}

// New config
func New(in []byte) (*Config, error) {
	c := &Config{}
	err := utils.UnmarshalYAML(in, c)
	return c, err
}
