package config

import (
	"time"

	"github.com/baetyl/baetyl-go/http"
	"github.com/baetyl/baetyl-go/log"
)

// Config the core config
type Config struct {
	Node struct {
		Name      string `json:"name,omitempty"`
		Namespace string `json:"namespace,omitempty"`
	} `yaml:"node" json:"node"`
	Engine EngineConfig `yaml:"engine" json:"engine"`
	Sync   SyncConfig   `yaml:"sync" json:"sync"`
	Store  StoreConfig  `yaml:"store" json:"store"`
	Logger log.Config   `yaml:"logger" json:"logger"`
}

type EngineConfig struct {
	Kind       string           `yaml:"kind" json:"kind" default:"kubernetes"`
	Kubernetes KubernetesConfig `yaml:"kubernetes" json:"kubernetes"`
	Collector  struct {
		Interval time.Duration `yaml:"interval" json:"interval"`
	} `yaml:"collector" json:"collector"`
}

type KubernetesConfig struct {
	InCluster  bool   `yaml:"inCluster" json:"inCluster" default:"false"`
	ConfigPath string `yaml:"configPath" json:"configPath"`
}

type StoreConfig struct {
	Path string `yaml:"path" json:"path" default:"var/lib/baetyl/store/core.db"`
}

type SyncConfig struct {
	Cloud struct {
		HTTP   http.ClientConfig `yaml:"http" json:"http"`
		Report struct {
			URL      string        `yaml:"url" json:"url" default:"/v1/sync/report"`
			Interval time.Duration `yaml:"interval" json:"interval" default:"10s"`
		} `yaml:"report" json:"report"`
		Desire struct {
			URL string `yaml:"url" json:"url" default:"/v1/sync/desire"`
		} `yaml:"desire" json:"desire"`
	} `yaml:"cloud" json:"cloud"`
	Edge struct {
		DownloadPath string `yaml:"downloadPath" json:"downloadPath" default:"var/lib/baetyl/download"`
	} `yaml:"edge" json:"edge"`
}
