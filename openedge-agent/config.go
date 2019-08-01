package main

import (
	"time"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/protocol/http"
	"github.com/baidu/openedge/protocol/mqtt"
)

// Config agent config
type Config struct {
	Remote struct {
		MQTT   mqtt.ClientInfo `yaml:"mqtt" json:"mqtt"`
		HTTP   http.ClientInfo `yaml:"http" json:"http"`
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
	Timeout time.Duration  `yaml:"timeout" json:"timeout" default:"10m"`
	Logger  logger.LogInfo `yaml:"logger" json:"logger" default:"{\"path\":\"var/db/openedge/openedge-log/openedge-ota.log\",\"format\":\"json\"}"`
}
