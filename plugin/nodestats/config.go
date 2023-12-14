// Package nodestats 节点状态监控实现
package nodestats

import "github.com/baetyl/baetyl-go/v2/http"

type Config struct {
	NodeStats struct {
		CollectPort       string `yaml:"collectPort" json:"collectPort" default:"30080"`
		CollectURL        string `yaml:"collectUrl" json:"collectUrl" default:"/v1/metrics"`
		Kube              KubeConfig
		http.ClientConfig `yaml:",inline" json:",inline"`
	} `yaml:"nodestats" json:"nodestats"`
}

type KubeConfig struct {
	OutCluster bool   `yaml:"outCluster" json:"outCluster"`
	ConfPath   string `yaml:"confPath" json:"confPath"`
}
