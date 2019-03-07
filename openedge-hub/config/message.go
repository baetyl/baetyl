package config

import (
	"time"

	"github.com/baidu/openedge/utils"
)

// Message message config
type Message struct {
	Length  utils.Length `yaml:"length" json:"length" default:"{\"max\":32768}"`
	Ingress struct {
		Qos0 struct {
			Buffer struct {
				Size int `yaml:"size" json:"size" default:"10000" validate:"min=1"`
			} `yaml:"buffer" json:"buffer"`
		} `yaml:"qos0" json:"qos0"`
		Qos1 struct {
			Buffer struct {
				Size int `yaml:"size" json:"size" default:"100" validate:"min=1"`
			} `yaml:"buffer" json:"buffer"`
			Batch struct {
				Max int `yaml:"max" json:"max" default:"50" validate:"min=1"`
			} `yaml:"batch" json:"batch"`
			Cleanup struct {
				Retention time.Duration `yaml:"retention" json:"retention" default:"48h"`
				Interval  time.Duration `yaml:"interval" json:"interval" default:"1m"`
			} `yaml:"cleanup" json:"cleanup"`
		} `yaml:"qos1" json:"qos1"`
	} `yaml:"ingress" json:"ingress"`
	Egress struct {
		Qos0 struct {
			Buffer struct {
				Size int `yaml:"size" json:"size" default:"10000" validate:"min=1"`
			} `yaml:"buffer" json:"buffer"`
		} `yaml:"qos0" json:"qos0"`
		Qos1 struct {
			Buffer struct {
				Size int `yaml:"size" json:"size" default:"100" validate:"min=1,max=65535"`
			} `yaml:"buffer" json:"buffer"`
			Batch struct {
				Max int `yaml:"max" json:"max" default:"50" validate:"min=1,max=10000"`
			} `yaml:"batch" json:"batch"`
			Retry struct {
				Interval time.Duration `yaml:"interval" json:"interval" default:"20s"`
			} `yaml:"retry" json:"retry"`
		} `yaml:"qos1" json:"qos1"`
	} `yaml:"egress" json:"egress"`
	Offset struct {
		Buffer struct {
			Size int `yaml:"size" json:"size" default:"10000" validate:"min=1"`
		} `yaml:"buffer" json:"buffer"`
		Batch struct {
			Max int `yaml:"max" json:"max" default:"100" validate:"min=1"`
		} `yaml:"batch" json:"batch"`
	} `yaml:"offset" json:"offset"`
}
