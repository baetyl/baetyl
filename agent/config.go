package agent

import "github.com/baidu/openedge/module/config"

// Config agent config
type Config struct {
	config.MQTTClient `yaml:",inline" json:",inline"`
	OpenAPI           config.HTTPClient `yaml:"openapi" json:"openapi"`
}
