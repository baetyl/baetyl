package main

import (
	"time"

	openedge "github.com/baidu/openedge/api/go"
)

// Config agent config
type Config struct {
	openedge.Config `yaml:",inline" json:",inline"`
	Remote          struct {
		MQTT   openedge.MqttClientInfo `yaml:"mqtt" json:"mqtt"`
		HTTP   openedge.HttpClientInfo `yaml:"http" json:"http"`
		Report struct {
			URL      string        `yaml:"url" json:"url" default:"v1/edge/info"`
			Topic    string        `yaml:"topic" json:"topic" default:"$baidu/iot/edge/%s/core/forward"`
			Interval time.Duration `yaml:"interval" json:"interval" default:"1m"`
		} `yaml:"report" json:"report"`
		Desire struct {
			Topic string `yaml:"topic" json:"topic" default:"$baidu/iot/edge/%s/core/backward"`
		} `yaml:"desire" json:"desire"`
	} `yaml:"remote" json:"remote"`
}
