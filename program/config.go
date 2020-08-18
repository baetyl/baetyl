package program

import "github.com/baetyl/baetyl-go/v2/log"

// Config is the program config.
type Config struct {
	Name        string `yaml:"name" json:"name" validate:"nonzero"`
	DisplayName string `yaml:"displayName" json:"displayName"`
	Description string `yaml:"description" json:"description"`

	Dir  string   `yaml:"dir" json:"dir"`
	Exec string   `yaml:"exec" json:"exec"`
	Args []string `yaml:"args" json:"args"`
	Env  []string `yaml:"env" json:"env"`

	Logger log.Config `yaml:"logger" json:"logger"`
}

type Entry struct {
	Entry string `yaml:"entry" json:"entry" validate:"nonzero"`
}
