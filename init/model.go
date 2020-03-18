package init

import (
	"github.com/baetyl/baetyl/sdk/baetyl-go"
)

type BackwardInfo struct {
	Data map[string]string `yaml:"data,omitempty" json:"data,omitempty"`
}

type ForwardInfo struct {
	Metadata map[string]string `yaml:"metadata" json:"metadata" default:"{}"`
	Status   baetyl.Inspect    `yaml:"status" json:"status"`
	Data     map[string]string `yaml:"data" json:"data" default:"{}"`
}
