package httplink

import "github.com/baetyl/baetyl-go/v2/http"

type Config struct {
	HTTPLink struct {
		HTTP      http.ClientConfig `yaml:"http" json:"http"`
		ReportURL string            `yaml:"reportUrl" json:"reportUrl" default:"v1/sync/report"`
		DesireURL string            `yaml:"desireUrl" json:"desireUrl" default:"v1/sync/desire"`
	} `yaml:"httplink" json:"httplink"`
}
