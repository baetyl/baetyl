package trans

import (
	"crypto/tls"

	"github.com/docker/go-connections/tlsconfig"
)

// NewTLSServerConfig loads tls config for server
func NewTLSServerConfig(caFile, keyFile, certFile string) (*tls.Config, error) {
	if certFile == "" && keyFile == "" {
		return nil, nil
	}
	return tlsconfig.Server(tlsconfig.Options{CAFile: caFile, KeyFile: keyFile, CertFile: certFile, ClientAuth: tls.VerifyClientCertIfGiven})
}

// NewTLSClientConfig loads tls config for client
func NewTLSClientConfig(caFile, keyFile, certFile string, insecureSkipVerify bool) (*tls.Config, error) {
	return tlsconfig.Client(tlsconfig.Options{CAFile: caFile, KeyFile: keyFile, CertFile: certFile, InsecureSkipVerify: insecureSkipVerify})
}
