package main

import (
	"github.com/baetyl/baetyl-go/link"
	"time"

	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/protocol/http"
	"github.com/baetyl/baetyl/protocol/mqtt"
	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
)

// Config agent config
type Config struct {
	Identity struct {
		Namespace string `yaml:"namespace" json:"namespace"`
		Name      string `yaml:"name" json:"name"`
	} `yaml:"identity" json:"identity"`
	Remote struct {
		MQTT   *mqtt.ClientInfo   `yaml:"mqtt" json:"mqtt"`
		HTTP   *http.ClientInfo   `yaml:"http" json:"http"`
		Link   *link.ClientConfig `yaml:"link" json:"link"`
		Report struct {
			URL      string        `yaml:"url" json:"url" default:"/v3/edge/info"`
			Topic    string        `yaml:"topic" json:"topic" default:"$baidu/iot/edge/%s/core/forward"`
			Interval time.Duration `yaml:"interval" json:"interval" default:"20s"`
		} `yaml:"report" json:"report"`
		Desire struct {
			Topic string `yaml:"topic" json:"topic" default:"$baidu/iot/edge/%s/core/backward"`
		} `yaml:"desire" json:"desire"`
	} `yaml:"remote" json:"remote"`
	OTA OTAInfo `yaml:"ota" json:"ota"`
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
	Request    map[string]string `yaml:"request" json:"request" default:"{}"`
	Status     *inspect          `yaml:"status" json:"status"`                      // node update
	DeployInfo map[string]string `yaml:"deployInfo" json:"deployInfo" default:"{}"` // shadow update
}

type BackwardInfo struct {
	Delta    map[string]interface{} `yaml:"delta" json:"delta"`
	Response map[string]interface{} `yaml:"response" json:"response"`
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
	ConfigMaps map[string]string `yaml:"configMaps" json:"configMaps" default:"{}"`
}

type ModuleConfig struct {
	Name      string            `yaml:"name" json:"name"`
	Namespace string            `yaml:"namespace" json:"namespace"`
	Data      map[string]string `yaml:"data" json:"data" default:"{}"`
	Version   string            `yaml:"version" json:"version"`
	Labels    map[string]string `yaml:"labels" json:"labels"`
}
