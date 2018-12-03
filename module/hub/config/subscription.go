package config

// Subscription subscription
type Subscription struct {
	Source struct {
		Topic string `yaml:"topic" json:"topic" validate:"nonzero"`
		QOS   byte   `yaml:"qos" json:"qos" default:"0" validate:"min=0, max=1"`
	} `yaml:"source" json:"source"`
	Target struct {
		Topic string `yaml:"topic" json:"topic" validate:"nonzero"`
		QOS   byte   `yaml:"qos" json:"qos" default:"0" validate:"min=0, max=1"`
	} `yaml:"target" json:"target"`
}
