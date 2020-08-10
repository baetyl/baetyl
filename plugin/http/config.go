package http

import "github.com/baetyl/baetyl-go/v2/http"

type Config struct {
	HTTPLink struct {
		HTTP   http.ClientConfig `yaml:"http" json:"http"`
		Report struct {
			URL      string        `yaml:"url" json:"url" default:"v1/sync/report"`
		} `yaml:"report" json:"report"`
		Desire struct {
			URL string `yaml:"url" json:"url" default:"v1/sync/desire"`
		} `yaml:"desire" json:"desire"`
	} `yaml:"httplink" json:"httplink"`
}
