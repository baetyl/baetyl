package main

import (
	"time"

	"github.com/baidu/openedge/module/config"
)

// Config agent config
type Config struct {
	config.Module `yaml:",inline" json:",inline"`
	API           config.HTTPClient `yaml:"api" json:"api"`
	Remote        struct {
		MQTT   config.MQTTClient `yaml:"mqtt" json:"mqtt"`
		HTTP   config.HTTPClient `yaml:"http" json:"http"`
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
