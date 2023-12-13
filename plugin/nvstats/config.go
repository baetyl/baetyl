// Package nvstats nv 状态监控实现
package nvstats

type Config struct {
	NvStats struct {
		CollectPort int32  `yaml:"collectPort" json:"collectPort" default:"30060"`
		CollectURL  string `yaml:"collectUrl" json:"collectUrl" default:"/v1/collect"`
		KubeConfig  `yaml:",inline" json:",inline"`
	} `yaml:"nvstats" json:"nvstats"`
}

type KubeConfig struct {
	OutCluster bool   `yaml:"outCluster" json:"outCluster"`
	ConfPath   string `yaml:"confPath" json:"confPath"`
}
