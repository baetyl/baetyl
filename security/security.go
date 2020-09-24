package security

import (
	"net"
	"net/url"

	"github.com/baetyl/baetyl-go/v2/pki"
)

//go:generate mockgen -destination=../mock/security.go -package=mock -source=security.go Security

// AltNames contains the domain names and IP addresses that will be added
// to the API Server's x509 certificate SubAltNames field. The values will
// be passed directly to the x509.Certificate object.
type AltNames struct {
	DNSNames []string   `json:"dnsNames,omitempty"`
	IPs      []net.IP   `json:"ips,omitempty"`
	Emails   []string   `json:"emails,omitempty"`
	URIs     []*url.URL `json:"uris,omitempty"`
}

type Security interface {
	// GetCA get self-signed root certificate crt
	GetCA() ([]byte, error)
	// IssueCertificate issuing sub-certificates through self-signed root certificates
	IssueCertificate(cn string, alt AltNames) (*pki.CertPem, error)
	// RevokeCertificate revoke certificate
	RevokeCertificate(cn string) error
	// RotateCertificate renew a certificate
	RotateCertificate(cn string) (*pki.CertPem, error)
}
