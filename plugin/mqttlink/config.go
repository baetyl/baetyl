// Package mqttlink 端云链接 mqtt 实现
package mqttlink

import (
	"github.com/baetyl/baetyl-go/v2/mqtt"
	"github.com/baetyl/baetyl-go/v2/utils"
)

type Config struct {
	MqttLink struct {
		mqtt.ClientConfig `yaml:",inline" json:",inline"`
		Report            mqtt.QOSTopic `yaml:"report" json:"report"`
		Delta             mqtt.QOSTopic `yaml:"delta" json:"delta"`
		Desire            mqtt.QOSTopic `yaml:"desire" json:"desire"`
		DesireResponse    mqtt.QOSTopic `yaml:"desireResponse" json:"desireResponse"`
	} `yaml:"mqttlink,omitempty" json:"mqttlink,omitempty"`
	Node utils.Certificate `yaml:"node" json:"node"`
}
