package main

import (
	"time"

	"github.com/baidu/openedge/protocol/http"
	"github.com/baidu/openedge/protocol/mqtt"
)

// Config agent config
type Config struct {
	Remote struct {
		MQTT   mqtt.ClientInfo `yaml:"mqtt" json:"mqtt"`
		HTTP   http.ClientInfo `yaml:"http" json:"http"`
		Report struct {
			URL      string        `yaml:"url" json:"url" default:"/v2/edge/info"`
			Topic    string        `yaml:"topic" json:"topic" default:"$baidu/iot/edge/%s/core/forward"`
			Interval time.Duration `yaml:"interval" json:"interval" default:"1m"`
		} `yaml:"report" json:"report"`
		Desire struct {
			Topic string `yaml:"topic" json:"topic" default:"$baidu/iot/edge/%s/core/backward"`
		} `yaml:"desire" json:"desire"`
		Meta struct {
			Cert string `yaml:"cert" json:"cert"`
		} `yaml:"meta" json:"meta"`
	} `yaml:"remote" json:"remote"`
}
