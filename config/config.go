package config

import (
	"github.com/baetyl/baetyl-go/v2/utils"
	"time"

	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/log"
)

// Config the core config
type Config struct {
	Engine   EngineConfig      `yaml:"engine" json:"engine"`
	Sync     SyncConfig        `yaml:"sync" json:"sync"`
	Cert     utils.Certificate `yaml:"cert" json:"cert"`
	Store    StoreConfig       `yaml:"store" json:"store"`
	Init     InitConfig        `yaml:"init" json:"init"`
	Security SecurityConfig    `yaml:"security" json:"security"`
	Server   http.ServerConfig `yaml:"server" json:"server"`
	Logger   log.Config        `yaml:"logger" json:"logger"`
	Plugin   struct {
		Link string `yaml:"link" json:"link" default:"httplink"`
	} `yaml:"plugin" json:"plugin"`
}

type AmiConfig struct {
	Kind       string           `yaml:"kind" json:"kind" default:"kubernetes"`
	Kubernetes KubernetesConfig `yaml:"kubernetes" json:"kubernetes"`
}

type EngineConfig struct {
	AmiConfig `yaml:",inline" json:",inline"`
	Report    struct {
		Interval time.Duration `yaml:"interval" json:"interval" default:"10s"`
	} `yaml:"report" json:"report"`
}

type KubernetesConfig struct {
	InCluster  bool                `yaml:"inCluster" json:"inCluster" default:"false"`
	ConfigPath string              `yaml:"configPath" json:"configPath"`
	LogConfig  KubernetesLogConfig `yaml:"logConfig" json:"logConfig"`
}

type KubernetesLogConfig struct {
	Follow     bool `yaml:"follow" json:"follow"`
	Previous   bool `yaml:"previous" json:"previous"`
	TimeStamps bool `yaml:"timestamps" json:"timestamps"`
}

type StoreConfig struct {
	Path string `yaml:"path" json:"path" default:"var/lib/baetyl/store/core.db"`
}

type SyncConfig struct {
	ReportInterval time.Duration `yaml:"reportInterval" json:"reportInterval" default:"20s"`
	Download       struct {
		Path              string `yaml:"path" json:"path" default:"var/lib/baetyl/download"`
		http.ClientConfig `yaml:",inline" json:",inline"`
	} `yaml:"download" json:"download"`
}

type InitConfig struct {
	Batch struct {
		Name         string `yaml:"name" json:"name"`
		Namespace    string `yaml:"namespace" json:"namespace"`
		SecurityType string `yaml:"securityType" json:"securityType"`
		SecurityKey  string `yaml:"securityKey" json:"securityKey"`
	} `yaml:"batch" json:"batch"`
	Active struct {
		http.ClientConfig `yaml:",inline" json:",inline"`
		URL               string        `yaml:"url" json:"url" default:"/v1/active"`
		Interval          time.Duration `yaml:"interval" json:"interval" default:"45s"`
	} `yaml:"active" json:"active"`
	ActivateConfig struct {
		Fingerprints []Fingerprint `yaml:"fingerprints" json:"fingerprints"`
		Attributes   []Attribute   `yaml:"attributes" json:"attributes"`
		Server       Server        `yaml:"server" json:"server"`
	} `yaml:"active" json:"active"`
}

type SecurityConfig struct {
	Kind      string    `yaml:"kind" json:"kind" default:"pki"`
	PKIConfig PKIConfig `yaml:"pki" json:"pki"`
}

type PKIConfig struct {
	SubDuration  time.Duration `yaml:"subDuration" json:"subDuration" default:"175200h"`   // 20*365*24
	RootDuration time.Duration `yaml:"rootDuration" json:"rootDuration" default:"438000h"` // 50*365*24
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
