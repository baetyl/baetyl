package main

import (
	"io/ioutil"
	"time"

	"github.com/baetyl/baetyl/protocol/mqtt"
	"github.com/docker/go-units"
	yaml "gopkg.in/yaml.v2"
)

// Kind the type of event from cloud
type Kind string

// The type of event from cloud
const (
	Bos  Kind = "BOS"
	Ceph Kind = "CEPH"
	S3   Kind = "S3"
)

// Config config of module
type Config struct {
	Clients []ClientInfo `yaml:"clients" json:"clients" default:"[]"`
	Rules   []RuleInfo   `yaml:"rules" json:"rules" default:"[]"`
}

// Retry policy
type Retry struct {
	Max   int           `yaml:"max" json:"max" default:"0"`       // retry max
	Delay time.Duration `yaml:"delay" json:"delay" default:"20s"` // delay time
	Base  time.Duration `yaml:"base" json:"base" default:"0.3s"`  // base time & *2
}

// Pool go pool
type Pool struct {
	Worker   int           `yaml:"worker" json:"worker" default:"1000"`    // max worker size
	Idletime time.Duration `yaml:"idletime" json:"idletime" default:"30s"` // delay time
}

// ClientInfo client config
type ClientInfo struct {
	Name      string        `yaml:"name" json:"name" validate:"nonzero"`
	Address   string        `yaml:"address" json:"address"`
	Region    string        `yaml:"region" json:"region" default:"us-east-1"`
	Ak        string        `yaml:"ak" json:"ak" validate:"nonzero"`
	Sk        string        `yaml:"sk" json:"sk" validate:"nonzero"`
	Kind      Kind          `yaml:"kind" json:"kind" validate:"nonzero"`
	Timeout   time.Duration `yaml:"timeout" json:"timeout" default:"30s"`
	Retry     Retry         `yaml:"retry" json:"retry"`
	Pool      Pool          `yaml:"pool" json:"pool"`
	Bucket    string        `yaml:"bucket" json:"bucket" validate:"nonzero"`
	TempPath  string        `yaml:"temppath" json:"temppath" default:"var/db/openedge/tmp"`
	MultiPart MultiPart     `yaml:"multipart" json:"multipart"`
	Limit     Limit         `yaml:"limit" json:"limit"`
	Report    struct {
		Interval time.Duration `yaml:"interval" json:"interval" default:"1m"`
	} `yaml:"report" json:"report"`
}

// RuleInfo function rule config
type RuleInfo struct {
	ClientID  string         `yaml:"clientid" json:"clientid"`
	Subscribe mqtt.TopicInfo `yaml:"subscribe" json:"subscribe"`
	Client    struct {
		Name string `yaml:"name" json:"name" validate:"nonzero"`
	} `yaml:"client" json:"client"`
}

// MultiPart config
type MultiPart struct {
	PartSize    int64 `yaml:"partsize" json:"partsize" default:"1048576000"`
	Concurrency int   `yaml:"concurrency" json:"concurrency" default:"10"`
}

type multipart struct {
	PartSize    string `yaml:"partsize" json:"partsize"`
	Concurrency int    `yaml:"concurrency" json:"concurrency"`
}

// Limit limit config
type Limit struct {
	Data int64  `yaml:"data" json:"data"`
	Path string `yaml:"path" json:"path" default:"var/db/openedge/data/stats.yml"`
}

type limit struct {
	Data string `yaml:"data" json:"data"`
	Path string `yaml:"path" json:"path"`
}

// UnmarshalYAML customizes unmarshal
func (l *Limit) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var ls limit
	err := unmarshal(&ls)
	if err != nil {
		return err
	}
	if ls.Data != "" {
		l.Data, err = units.RAMInBytes(ls.Data)
		if err != nil {
			return err
		}
	}
	if ls.Path != "" {
		l.Path = ls.Path
	}
	return nil
}

// UnmarshalYAML customizes unmarshal
func (m *MultiPart) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var ms multipart
	err := unmarshal(&ms)
	if err != nil {
		return err
	}
	if ms.PartSize != "" {
		m.PartSize, err = units.RAMInBytes(ms.PartSize)
		if err != nil {
			return err
		}
	}
	if ms.Concurrency != 0 {
		m.Concurrency = ms.Concurrency
	}
	return nil
}

// Item data count
type Item struct {
	Bytes int64 `yaml:"bytes" json:"bytes" default:"0"`
	Count int64 `yaml:"count" json:"count" default:"0"`
}

// Stats month stats
type Stats struct {
	Total  Item             `yaml:"total" json:"total" default:"{}"`
	Months map[string]*Item `yaml:"months" json:"months" default:"{}"`
}

// DumpYAML in interface save in config file
func DumpYAML(path string, in interface{}) error {
	bytes, err := yaml.Marshal(in)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, bytes, 0755)
	if err != nil {
		return err
	}
	return nil
}
