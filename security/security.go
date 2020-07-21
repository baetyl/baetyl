package security

import (
	"net"
	"net/url"
)

//go:generate mockgen -destination=../mock/security.go -package=mock github.com/baetyl/baetyl/security Security

// AltNames contains the domain names and IP addresses that will be added
// to the API Server's x509 certificate SubAltNames field. The values will
// be passed directly to the x509.Certificate object.
type AltNames struct {
	DNSNames []string   `json:"dnsNames,omitempty"`
	IPs      []net.IP   `json:"ips,omitempty"`
	Emails   []string   `json:"emails,omitempty"`
	URIs     []*url.URL `json:"uris,omitempty"`
}

// PEMCredential holds a certificate, private key pem data and certificate ID
type PEMCredential struct {
	CertPEM []byte
	KeyPEM  []byte
	CertId  string
}

type Security interface {
	GetCA() ([]byte, error)
	IssueCertificate(cn string, alt AltNames) (*PEMCredential, error)
	RevokeCertificate(certId string) error
	RotateCertificate(certId string) (*PEMCredential, error)
}
