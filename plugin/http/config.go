package http

import "github.com/baetyl/baetyl-go/v2/http"

type Config struct {
	HTTP http.ClientConfig `yaml:"http" json:"http"`
}
