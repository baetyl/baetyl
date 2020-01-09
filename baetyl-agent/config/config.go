package config

import (
	"time"

	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/protocol/http"
	"github.com/baetyl/baetyl/protocol/mqtt"
	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
)

// Config agent config
type Config struct {
	Remote struct {
		MQTT   *mqtt.ClientInfo   `yaml:"mqtt" json:"mqtt"`
		HTTP   *http.ClientInfo   `yaml:"http" json:"http" default:"{}"`
		Report struct {
			URL      string        `yaml:"url" json:"url" default:"/v3/edge/info"`
			Topic    string        `yaml:"topic" json:"topic" default:"$baidu/iot/edge/%s/core/forward"`
			Interval time.Duration `yaml:"interval" json:"interval" default:"20s"`
		} `yaml:"report" json:"report"`
		Desire struct {
			URL string `yaml:"url" json:"url"`
			Topic string `yaml:"topic" json:"topic" default:"$baidu/iot/edge/%s/core/backward"`
		} `yaml:"desire" json:"desire"`
	} `yaml:"remote" json:"remote"`
	OTA    OTAInfo `yaml:"ota" json:"ota"`
	Active `yaml:",inline" json:",inline"`
}

// OTAInfo ota config
type OTAInfo struct {
	Timeout time.Duration  `yaml:"timeout" json:"timeout" default:"5m"`
	Logger  logger.LogInfo `yaml:"logger" json:"logger" default:"{\"path\":\"var/db/baetyl/volumes/ota.log\",\"format\":\"json\"}"`
}

// Metadata meta data of volume
type Metadata struct {
	Version string              `yaml:"version" json:"version"`
	Volumes []baetyl.VolumeInfo `yaml:"volumes" json:"volumes" default:"[]"`
}

type DeployConfig struct {
	AppConfig baetyl.ComposeAppConfig      `yaml:"appConfig" json:"appConfig"`
	Metadata  map[string]baetyl.VolumeInfo `yaml:"metadata" json:"metadata" default:"{}"`
}

type ForwardInfo struct {
	Namespace  string            `yaml:"namespace" json:"namespace"`
	Name       string            `yaml:"name" json:"mame"`
	Metadata   map[string]string `yaml:"metadata" json:"metadata" default:"{}"`
	Status     *Inspect          `yaml:"status" json:"status"`                      // node update
	Deployment map[string]string `yaml:"deployment" json:"deployment" default:"{}"` // shadow update
	Activation Activation        `yaml:"activation" json:"activation"`
}

type Activation struct {
	FingerprintValue string            `yaml:"fingerprintValue" json:"fingerprintValue" default:""`
	PenetrateData    map[string]string `yaml:"penetrateData" json:"penetrateData" default:"{}"`
}

type BackwardInfo struct {
	Delta    map[string]interface{} `yaml:"delta,omitempty" json:"delta,omitempty"`
	Metadata map[string]interface{} `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

type Deployment struct {
	Name      string `yaml:"name" json:"name"`
	Namespace string `yaml:"namespace" json:"namespace"`
	Version   string `yaml:"version" json:"version"`
	Selector  string `yaml:"selector" json:"selector"`
	// Snapshot for app and config
	Snapshot    Snapshot `yaml:"snapshot" json:"snapshot"`
	Priority    int      `yaml:"priority" json:"priority"`
	Description string   `yaml:"description" json:"description"`
}

type Snapshot struct {
	// key = unique name of the app
	// value = version of the app
	Apps map[string]string `yaml:"apps" json:"apps" default:"{}"`
	// key = unique name of the volume
	// value = version of the volume
	Configs map[string]string `yaml:"configs" json:"configs" default:"{}"`
}

type ModuleConfig struct {
	Name      string            `yaml:"name" json:"name"`
	Namespace string            `yaml:"namespace" json:"namespace"`
	Data      map[string]string `yaml:"data" json:"data" default:"{}"`
	Version   string            `yaml:"version" json:"version"`
	Labels    map[string]string `yaml:"labels" json:"labels"`
}

// Active
// Config active module config
type Active struct {
	Interval     time.Duration `yaml:"interval" json:"interval" default:"1m"  validate:"min=10000000000"`
	Fingerprints []Fingerprint `yaml:"fingerprints" json:"fingerprints"`
	Attributes   []Attribute   `yaml:"attributes" json:"attributes"`
	Server       Server        `yaml:"server" json:"server"`
}

// Server manually activated server configuration
type Server struct {
	Listen string `yaml:"listen" json:"listen"`
	Pages  string `yaml:"pages" json:"pages" default:"etc/baetyl/pages"`
}

// Fingerprint type to be collected
type Fingerprint struct {
	Proof string `yaml:"proof" json:"proof"`
	Value string `yaml:"value" json:"value"`
}

// Attribute field to be filled (masterKey, deviceType, deviceCompany)
type Attribute struct {
	Name  string `yaml:"name" json:"name" validate:"nonzero"`
	Label string `yaml:"label" json:"label" validate:"nonzero"`
	Value string `yaml:"value" json:"value"`
	Desc  string `yaml:"description" json:"description"`
}

type Inspect struct {
	*baetyl.Inspect `json:",inline"`
	OTA             map[string][]*Record `json:"ota,omitempty"`
}

type Record struct {
	Time  string `json:"time,omitempty"`
	Step  string `json:"step,omitempty"`
	Trace string `json:"trace,omitempty"`
	Error string `json:"error,omitempty"`
}

type ResourceRequest struct {
	Type string `yaml:"type" json:"type"`
	Name string `yaml:"name" json:"name"`
	Version string `yaml:"version" json:"version"`
}

type ResourceResponse struct {
	Deployment Deployment `yaml:"deployment" json:"deployment"`
	Application DeployConfig `yaml:"application" json:"application"`
	Configs map[string]ModuleConfig `yaml:"configs" json:"configs"`
}
