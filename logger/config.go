package logger

// Config logger config
type Config struct {
	Path    string `yaml:"path" json:"path"`
	Level   string `yaml:"level" json:"level" default:"info" validate:"regexp=^(info|debug|warn|error)$"`
	Format  string `yaml:"format" json:"format" default:"text" validate:"regexp=^(text|json)$"`
	Console bool   `yaml:"console" json:"console" default:"false"`
	Age     struct {
		Max int `yaml:"max" json:"max" default:"15" validate:"min=1"`
	} `yaml:"age" json:"age"` // days
	Size struct {
		Max int `yaml:"max" json:"max" default:"50" validate:"min=1"`
	} `yaml:"size" json:"size"` // in MB
	Backup struct {
		Max int `yaml:"max" json:"max" default:"15" validate:"min=0"`
	} `yaml:"backup" json:"backup"`
}
