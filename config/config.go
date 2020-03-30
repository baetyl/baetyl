package config

import (
	"time"

	"github.com/baetyl/baetyl-go/http"
	"github.com/baetyl/baetyl-go/log"
)

// Config the core config
type Config struct {
	Engine EngineConfig `yaml:"engine" json:"engine"`
	Sync   SyncConfig   `yaml:"sync" json:"sync"`
	Store  StoreConfig  `yaml:"store" json:"store"`
	Init   InitConfig   `yaml:"init" json:"init"`
	Logger log.Config   `yaml:"logger" json:"logger"`
}

type EngineConfig struct {
	Kind       string           `yaml:"kind" json:"kind" default:"kubernetes"`
	Kubernetes KubernetesConfig `yaml:"kubernetes" json:"kubernetes"`
	Collector  struct {
		Interval time.Duration `yaml:"interval" json:"interval" default:"20s"`
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
	Node struct {
		Name      string `json:"name,omitempty"`
		Namespace string `json:"namespace,omitempty"`
	} `yaml:"node" json:"node"`
	Cloud struct {
		HTTP   http.ClientConfig `yaml:"http" json:"http"`
		Report struct {
			URL string `yaml:"url" json:"url" default:"v1/sync/report"`
		} `yaml:"report" json:"report"`
		Desire struct {
			URL string `yaml:"url" json:"url" default:"v1/sync/desire"`
		} `yaml:"desire" json:"desire"`
	} `yaml:"cloud" json:"cloud"`
	Edge struct {
		DownloadPath string `yaml:"downloadPath" json:"downloadPath" default:"var/lib/baetyl/download"`
	} `yaml:"edge" json:"edge"`
}

type InitConfig struct {
	Batch struct {
		Name         string `yaml:"name" json:"name"`
		Namespace    string `yaml:"namespace" json:"namespace"`
		SecurityType string `yaml:"securityType" json:"securityType"`
		SecurityKey  string `yaml:"securityKey" json:"securityKey"`
	} `yaml:"batch" json:"batch"`
	Cloud struct {
		HTTP   http.ClientConfig `yaml:"http" json:"http"`
		Active struct {
			URL      string        `yaml:"url" json:"url" default:"/v1/active"`
			Interval time.Duration `yaml:"interval" json:"interval" default:"45s"`
		} `yaml:"active" json:"active"`
	} `yaml:"cloud" json:"cloud"`
	Edge struct {
		DownloadPath string `yaml:"downloadPath" json:"downloadPath" default:"var/lib/baetyl/download"`
	} `yaml:"edge" json:"edge"`
	ActivateConfig struct {
		Fingerprints []Fingerprint `yaml:"fingerprints" json:"fingerprints"`
		Attributes   []Attribute   `yaml:"attributes" json:"attributes"`
		Server       Server        `yaml:"server" json:"server"`
	} `yaml:"active" json:"active"`
}

// Server manually activated server configuration
type Server struct {
	Listen string `yaml:"listen" json:"listen"`
	Pages  string `yaml:"pages" json:"pages" default:"etc/baetyl/pages"`
}

// Fingerprint type to be collected
type Fingerprint struct {
	Proof Proof  `yaml:"proof" json:"proof"`
	Value string `yaml:"value" json:"value"`
}

// Attribute field to be filled
type Attribute struct {
	Name  string `yaml:"name" json:"name" validate:"nonzero"`
	Label string `yaml:"label" json:"label" validate:"nonzero"`
	Value string `yaml:"value" json:"value"`
	Desc  string `yaml:"description" json:"description"`
}

// Proof the proof of fingerprints
type Proof string

// all proofs
const (
	ProofSN         Proof = "sn"
	ProofInput      Proof = "input"
	ProofHostName   Proof = "hostName"
	ProofBootID     Proof = "bootID"
	ProofMachineID  Proof = "machineID"
	ProofSystemUUID Proof = "systemUUID"
)
