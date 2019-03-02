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

// DatasetInfo dataset information
type DatasetInfo struct {
	Name    string `yaml:"name" json:"name" validate:"regexp=^[a-zA-Z0-9-_]{1,32}$"`
	Version string `yaml:"version" json:"version" validate:"regexp=^[a-zA-Z0-9-_]{0,32}$"`
	URL     string `yaml:"url" json:"url"`
	MD5     string `yaml:"md5" json:"md5"`
}

// UpdateEvent update event
type UpdateEvent struct {
	Version  string                 `yaml:"version" json:"version" validate:"regexp=^[a-zA-Z0-9-_]{1,32}$"`
	Config   openedge.AppConfig `yaml:"config" json:"config" validate:"nonzero"`
	Datasets []DatasetInfo          `yaml:"datasets" json:"datasets" default:"[]"`
}

func newUpdateEvent(d []byte) (*UpdateEvent, error) {
	data := new(UpdateEvent)
	err := utils.UnmarshalJSON(d, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
