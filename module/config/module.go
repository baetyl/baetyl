package config

// Module module config
type Module struct {
	Name   string `yaml:"name" json:"name" validate:"nonzero"`
	// Mark   string `yaml:"mark" json:"mark"`
	Logger Logger `yaml:"logger" json:"logger"`

	Entry     string            `yaml:"entry" json:"entry"`
	Restart   Policy            `yaml:"restart" json:"restart"`
	Expose    []string          `yaml:"expose" json:"expose" default:"[]"`
	Params    []string          `yaml:"params" json:"params" default:"[]"`
	Env       map[string]string `yaml:"env" json:"env" default:"{}"`
	Resources Resources         `yaml:"resources" json:"resources"`
}
