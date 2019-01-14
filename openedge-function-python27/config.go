package main

import openedge "github.com/baidu/openedge/api/go"

// Config of runtime
type Config struct {
	openedge.Config `yaml:",inline" json:",inline"`
	Subscribe       openedge.TopicInfo `yaml:"subscribe" json:"subscribe"`
	Publish         openedge.TopicInfo `yaml:"publish" json:"publish"`
	Name            string             `yaml:"name" json:"name" validate:"nonzero"`
	Handler         string             `yaml:"handler" json:"handler" validate:"nonzero"`
}
