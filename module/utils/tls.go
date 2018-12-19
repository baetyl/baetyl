package utils

import (
	"crypto/tls"

	"github.com/docker/go-connections/tlsconfig"
)

// Certificate certificate config for mqtt server
type Certificate struct {
	CA       string `yaml:"ca" json:"ca"`
	Key      string `yaml:"key" json:"key"`
	Cert     string `yaml:"cert" json:"cert"`
	Insecure bool   `yaml:"insecure" json:"insecure"` // for client, for test purpose
}

// NewTLSServerConfig loads tls config for server
func NewTLSServerConfig(c Certificate) (*tls.Config, error) {
	if c.Cert == "" && c.Key == "" {
		return nil, nil
	}
	return tlsconfig.Server(tlsconfig.Options{CAFile: c.CA, KeyFile: c.Key, CertFile: c.Cert, ClientAuth: tls.VerifyClientCertIfGiven})
}

// NewTLSClientConfig loads tls config for client
func NewTLSClientConfig(c Certificate) (*tls.Config, error) {
	return tlsconfig.Client(tlsconfig.Options{CAFile: c.CA, KeyFile: c.Key, CertFile: c.Cert, InsecureSkipVerify: c.Insecure})
}
