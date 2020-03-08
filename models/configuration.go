package models

type Configuration struct {
	Name      string            `yaml:"name" json:"name"`
	Namespace string            `yaml:"namespace" json:"namespace"`
	Data      map[string]string `yaml:"data" json:"data" default:"{}"`
	Version   string            `yaml:"version" json:"version"`
	Labels    map[string]string `yaml:"labels" json:"labels"`
}
