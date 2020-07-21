package security

import (
	"net"
	"net/url"
	"os"
	"sync"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl/config"
	bh "github.com/timshannon/bolthold"
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

const (
	PKI = "pki"
)

var mu sync.Mutex
var securityNews = map[string]New{}
var securityImpls = map[string]Security{}

type New func(cfg config.SecurityConfig, bhSto *bh.Store) (Security, error)

type Security interface {
	GetCA() ([]byte, error)
	IssueCertificate(cn string, alt AltNames) (*PEMCredential, error)
	RevokeCertificate(certId string) error
	RotateCertificate(certId string) (*PEMCredential, error)
}

func NewSecurity(cfg config.SecurityConfig, bhSto *bh.Store) (Security, error) {
	name := cfg.Kind
	mu.Lock()
	defer mu.Unlock()
	if sec, ok := securityImpls[name]; ok {
		return sec, nil
	}
	secNew, ok := securityNews[name]
	if !ok {
		return nil, errors.Trace(os.ErrInvalid)
	}
	sec, err := secNew(cfg, bhSto)
	if err != nil {
		return nil, errors.Trace(err)
	}
	securityImpls[name] = sec
	return sec, nil
}

func Register(name string, n New) {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := securityNews[name]; ok {
		log.L().Info("ami generator already exists, skip", log.Any("generator", name))
		return
	}
	log.L().Info("ami generator registered", log.Any("generator", name))
	securityNews[name] = n
}
