package config

import (
	"time"

	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/utils"
)

// Config the core config
type Config struct {
	Node     utils.Certificate `yaml:"node" json:"node"`
	Engine   EngineConfig      `yaml:"engine" json:"engine"`
	AMI      AmiConfig         `yaml:",inline" json:",inline"`
	Sync     SyncConfig        `yaml:"sync" json:"sync"`
	Store    StoreConfig       `yaml:"store" json:"store"`
	Event    EventConfig       `yaml:"event" json:"event"`
	Init     InitConfig        `yaml:"init" json:"init"`
	Security SecurityConfig    `yaml:"security" json:"security"`
	Server   http.ServerConfig `yaml:"server" json:"server"`
	Logger   log.Config        `yaml:"logger" json:"logger"`
	Plugin   struct {
		Link   string `yaml:"link" json:"link" default:"httplink"`
		Pubsub string `yaml:"pubsub" json:"pubsub" default:"defaultpubsub"`
	} `yaml:"plugin" json:"plugin"`
	MqttLink struct {
		Cert utils.Certificate `yaml:",inline" json:",inline"`
	} `yaml:"mqttlink,omitempty" json:"mqttlink,omitempty"`
}

type EngineConfig struct {
	Report struct {
		Interval time.Duration `yaml:"interval" json:"interval" default:"10s"`
	} `yaml:"report" json:"report"`
	Clean struct {
		Interval time.Duration `yaml:"interval" json:"interval" default:"10m"`
	} `yaml:"clean" json:"clean"`
}

type EventConfig struct {
	Notify  bool    `yaml:"notify" json:"notify"`
	Publish Publish `yaml:"publish" json:"publish"`
}

type Publish struct {
	QOS   byte   `yaml:"qos" json:"qos" validate:"min=0,max=1"`
	Topic string `yaml:"topic" json:"topic" default:"$baetyl/node/props" validate:"nonzero"`
}

type AmiConfig struct {
	Kube   KubeConfig   `yaml:"kube" json:"kube"`
	Native NativeConfig `yaml:"native" json:"native"`
}

type KubeConfig struct {
	OutCluster bool                `yaml:"outCluster" json:"outCluster"`
	ConfPath   string              `yaml:"confPath" json:"confPath"`
	LogConfig  KubernetesLogConfig `yaml:"logConfig" json:"logConfig"` // TODO: remove
}

type NativeConfig struct {
	PortsRange PortsRange `yaml:"portsRange" json:"portsRange"`
}

type PortsRange struct {
	Start int `yaml:"start" json:"start" default:"50200"`
	End   int `yaml:"end" json:"end" default:"51000"`
}

// TODO: remove
type KubernetesLogConfig struct {
	Follow     bool `yaml:"follow" json:"follow"`
	Previous   bool `yaml:"previous" json:"previous"`
	TimeStamps bool `yaml:"timestamps" json:"timestamps"`
}

type StoreConfig struct {
	Path string `yaml:"path" json:"path" default:"var/lib/baetyl/store/core.db"`
}

type SyncConfig struct {
	Report struct {
		Interval time.Duration `yaml:"interval" json:"interval" default:"20s"`
	} `yaml:"report" json:"report"`
	Download struct {
		Path              string `yaml:"path" json:"path" default:"var/lib/baetyl/object"`
		http.ClientConfig `yaml:",inline" json:",inline"`
	} `yaml:"download" json:"download"`
}

type InitConfig struct {
	Batch struct {
		Name         string `yaml:"name" json:"name"`
		Namespace    string `yaml:"namespace" json:"namespace"`
		SecurityType string `yaml:"securityType" json:"securityType"`
		SecurityKey  string `yaml:"securityKey" json:"securityKey"`
		Mode         string `yaml:"mode" json:"mode" default:"kube"`
	} `yaml:"batch" json:"batch"`
	Active struct {
		http.ClientConfig `yaml:",inline" json:",inline"`
		URL               string        `yaml:"url" json:"url" default:"/v1/active"`
		Interval          time.Duration `yaml:"interval" json:"interval" default:"45s"`
		Collector         struct {
			Fingerprints []Fingerprint `yaml:"fingerprints" json:"fingerprints"`
			Attributes   []Attribute   `yaml:"attributes" json:"attributes"`
			NodeInfo     []Attribute   `yaml:"nodeinfo" json:"nodeinfo"`
			Serial       []Attribute   `yaml:"serial" json:"serial"`
			Server       Server        `yaml:"server" json:"server"`
		} `yaml:"collector" json:"collector"`
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
