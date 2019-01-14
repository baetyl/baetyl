package config

// Module module config
type Module struct {
	Name   string `yaml:"name" json:"name" validate:"nonzero"`
	Alias  string `yaml:"alias" json:"alias"`
	Logger Logger `yaml:"logger" json:"logger"`

	Entry     string            `yaml:"entry" json:"entry"`
	Restart   Policy            `yaml:"restart" json:"restart"`
	Expose    []string          `yaml:"expose" json:"expose" default:"[]"`
	Volumes   []string          `yaml:"volumes" json:"volumes" default:"[]"`
	Params    []string          `yaml:"params" json:"params" default:"[]"`
	Env       map[string]string `yaml:"env" json:"env" default:"{}"`
	Resources Resources         `yaml:"resources" json:"resources"`
}

// UniqueName unique name of module
func (m *Module) UniqueName() string {
	if m.Alias == "" {
		return m.Name
	}
	return m.Alias
}
