package logger

// LogInfo for logging
type LogInfo struct {
	Path    string `yaml:"path" json:"path"`
	Level   string `yaml:"level" json:"level" default:"info" validate:"regexp=^(info|debug|warn|error)$"`
	Format  string `yaml:"format" json:"format" default:"text" validate:"regexp=^(text|json)$"`
	Age     struct {
		Max int `yaml:"max" json:"max" default:"15" validate:"min=1"`
	} `yaml:"age" json:"age"` // days
	Size struct {
		Max int `yaml:"max" json:"max" default:"50" validate:"min=1"`
	} `yaml:"size" json:"size"` // in MB
	Backup struct {
		Max int `yaml:"max" json:"max" default:"15" validate:"min=1"`
	} `yaml:"backup" json:"backup"`
}
