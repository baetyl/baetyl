// Package ws 实现端云基于ws协议的链接
package ws

import (
	"time"

	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/utils"
)

type Config struct {
	WSLink struct {
		http.ClientConfig    `yaml:",inline" json:",inline"`
		SyncURL              string        `yaml:"syncUrl" json:"syncUrl" default:"v1/sync"`
		ReconnectBackoff     Backoff       `yaml:"reconnectBackoff" json:"reconnectBackoff" default:"{\"min\":1000000000,\"max\":60000000000,\"factor\":2}"`
		WaitResponseInterval time.Duration `yaml:"waitResponseInterval" json:"waitResponseInterval" default:"10s"`
	} `yaml:"wslink" json:"wslink"`
	Node utils.Certificate `yaml:"node" json:"node"`
	Sync SyncConfig        `yaml:"sync" json:"sync"`
}

type Backoff struct {
	Min    time.Duration `yaml:"min" json:"min"`
	Max    time.Duration `yaml:"max" json:"max"`
	Factor float64       `yaml:"factor" json:"factor"`
}

type SyncConfig struct {
	Report struct {
		Interval time.Duration `yaml:"interval" json:"interval" default:"20s"`
	} `yaml:"report" json:"report"`
}
