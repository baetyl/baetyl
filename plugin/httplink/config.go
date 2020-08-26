package httplink

import (
	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/utils"
)

type Config struct {
	HTTPLink struct {
		HTTP      http.ClientConfig `yaml:",inline" json:",inline"`
		ReportURL string            `yaml:"reportUrl" json:"reportUrl" default:"v1/sync/report"`
		DesireURL string            `yaml:"desireUrl" json:"desireUrl" default:"v1/sync/desire"`
	} `yaml:"httplink" json:"httplink"`
	Node utils.Certificate `yaml:"node" json:"node"`
}
