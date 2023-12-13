// Package qpsstats qps监控实现
package qpsstats

import "github.com/baetyl/baetyl-go/v2/http"

type Config struct {
	QPSStats struct {
		CollectPort       string     `yaml:"collectPort" json:"collectPort" default:"30080"`
		CollectURL        string     `yaml:"collectUrl" json:"collectUrl" default:"status/format/json"`
		Kube              KubeConfig `yaml:"kube" json:"kube"`
		http.ClientConfig `yaml:",inline" json:",inline"`
	} `yaml:"qpsstats" json:"qpsstats"`
}

type KubeConfig struct {
	OutCluster bool   `yaml:"outCluster" json:"outCluster"`
	ConfPath   string `yaml:"confPath" json:"confPath"`
}
