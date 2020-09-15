package memory

import "time"

type CloudConfig struct {
	MQ struct {
		Size     int           `yaml:"size" json:"size" default:"10"`
		Duration time.Duration `yaml:"duration" json:"duration" default:"10m"`
	} `yaml:"defaultmq" json:"defaultmq"`
}
