package config

import (
	"time"

	"github.com/256dpi/gomqtt/packet"
	"github.com/baidu/openedge/module/utils"
)

// Subscription is a single subscription in a Subscribe packet.
type Subscription struct {
	Topic string     `yaml:"topic" json:"topic"`
	QOS   packet.QOS `yaml:"qos" json:"qos" default:"0"`
}

// MQTTClient mqtt client config
type MQTTClient struct {
	Address       string         `yaml:"address" json:"address"`
	ClientID      string         `yaml:"clientid" json:"clientid"`
	CleanSession  bool           `yaml:"cleansession" json:"cleansession" default:"false"`
	Timeout       time.Duration  `yaml:"timeout" json:"timeout" default:"30s"`
	Interval      time.Duration  `yaml:"interval" json:"interval" default:"1m"`
	KeepAlive     time.Duration  `yaml:"keepalive" json:"keepalive" default:"1m"`
	BufferSize    int            `yaml:"buffersize" json:"buffersize" default:"10"`
	ValidateSubs  bool           `yaml:"validatesubs" json:"validatesubs"`
	Subscriptions []Subscription `yaml:"subscriptions" json:"subscriptions" default:"[]"`

	Username          string `yaml:"username" json:"username"`
	Password          string `yaml:"password" json:"password"`
	utils.Certificate `yaml:",inline" json:",inline"`
}

// GetSubscriptions returns packet subscriptions
func (c *MQTTClient) GetSubscriptions() []packet.Subscription {
	output := make([]packet.Subscription, len(c.Subscriptions))
	for i, s := range c.Subscriptions {
		output[i] = packet.Subscription{Topic: s.Topic, QOS: s.QOS}
	}
	return output
}
