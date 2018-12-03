package config

// Storage storage config
type Storage struct {
	Dir string `yaml:"dir" json:"dir" default:"var/db"`
}
