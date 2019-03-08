package main

import (
	"time"

	"github.com/baidu/openedge/protocol/http"
	"github.com/baidu/openedge/protocol/mqtt"
	"github.com/baidu/openedge/sdk-go/openedge"
	"github.com/baidu/openedge/utils"
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
	} `yaml:"remote" json:"remote"`
}

// UpdateEvent update event
type UpdateEvent struct {
	openedge.AppConfig `yaml:",inline" json:",inline"`
}

func newUpdateEvent(d []byte) (*UpdateEvent, error) {
	data := new(UpdateEvent)
	err := utils.UnmarshalJSON(d, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
