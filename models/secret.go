package models

type Secret struct {
	Name      string            `yaml:"name" json:"name"`
	Namespace string            `yaml:"namespace" json:"namespace"`
	Data      map[string][]byte `yaml:"data" json:"data" default:"{}"`
	Version   string            `yaml:"version" json:"version"`
	Labels    map[string]string `yaml:"labels" json:"labels"`
}
