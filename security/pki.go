package security

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"net"
	"net/url"
	"os"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/pki"
	bh "github.com/timshannon/bolthold"

	"github.com/baetyl/baetyl/v2/config"
	"github.com/baetyl/baetyl/v2/sync"
)

type pkiClient struct {
	cfg config.SecurityConfig
	cli pki.PKI
	sto *bh.Store
	log *log.Logger
}

func NewPKI(cfg config.SecurityConfig, sto *bh.Store) (Security, error) {
	cli, err := pki.NewPKIClient()
	if err != nil {
		return nil, errors.Trace(err)
	}
	defaultCli := &pkiClient{
		cfg: cfg,
		cli: cli,
		sto: sto,
		log: log.With(log.Any("security", cfg.Kind)),
	}

	err = defaultCli.genSelfSignedCACertificate()
	if err != nil {
		return nil, errors.Trace(err)
	}
	return defaultCli, nil
}

func (p *pkiClient) GetCA() ([]byte, error) {
	cert, err := p.getCA()
	if err != nil {
		return nil, errors.Trace(err)
	}
	return cert.Crt, nil
}

func (p *pkiClient) IssueCertificate(cn string, alt AltNames) (*pki.CertPem, error) {
	ca, err := p.getCA()
	if err != nil {
		return nil, errors.Trace(err)
	}

	return p.cli.CreateSubCertWithKey(genCsr(cn, alt), (int)(p.cfg.PKIConfig.SubDuration.Hours()/24), ca)
}

func (p *pkiClient) RevokeCertificate(cn string) error {
	tp := pki.CertPem{}
	err := p.sto.Delete(genStoKey(cn), tp)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (p *pkiClient) RotateCertificate(cn string) (*pki.CertPem, error) {
	key := genStoKey(cn)
	cert := &pki.CertPem{}
	err := p.sto.Get(key, cert)
	if err != nil {
		return nil, errors.Trace(err)
	}
	err = p.RevokeCertificate(cn)
	if err != nil {
		return nil, errors.Trace(err)
	}
	certInfo, err := pki.ParseCertificates(cert.Crt)
	if err != nil {
		return nil, errors.Trace(err)
	}
	if len(certInfo) != 1 {
		return nil, errors.Trace(errors.Errorf("rotate certificate error"))
	}

	alt := AltNames{
		DNSNames: certInfo[0].DNSNames,
		IPs:      certInfo[0].IPAddresses,
		Emails:   certInfo[0].EmailAddresses,
		URIs:     certInfo[0].URIs,
	}
	return p.IssueCertificate(certInfo[0].Subject.CommonName, alt)
}

func (p *pkiClient) genSelfSignedCACertificate() error {
	cn := fmt.Sprintf("%s.%s", os.Getenv(sync.EnvKeyNodeNamespace), os.Getenv(sync.EnvKeyNodeName))
	csrInfo := genCsr(cn, AltNames{
		IPs: []net.IP{
			net.IPv4(0, 0, 0, 0),
			net.IPv4(127, 0, 0, 1),
		},
		URIs: []*url.URL{
			{
				Scheme: "https",
				Host:   "localhost",
			},
		},
	})
	cert, err := p.cli.CreateSelfSignedRootCert(csrInfo, (int)(p.cfg.PKIConfig.RootDuration.Hours()/24))
	if err != nil {
		return errors.Trace(err)
	}
	// save root ca
	err = p.putCert(genStoKey(cn), *cert)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (p *pkiClient) getCA() (*pki.CertPem, error) {
	cn := fmt.Sprintf("%s.%s", os.Getenv(sync.EnvKeyNodeNamespace), os.Getenv(sync.EnvKeyNodeName))
	key := genStoKey(cn)
	cert := &pki.CertPem{}
	err := p.sto.Get(key, cert)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return cert, nil
}

func (p *pkiClient) putCert(key string, cert pki.CertPem) error {
	err := p.sto.Insert(key, cert)
	if err != nil && err.Error() != bh.ErrKeyExists.Error() {
		return errors.Trace(err)
	}
	if err != nil {
		p.log.Info("baetyl internal certificate already exists.", log.Any("key", key))
	}
	return nil
}

func genStoKey(cn string) string {
	return fmt.Sprintf("%s-%s", "baetyl-cert", cn)
}

func genCsr(cn string, alt AltNames) *x509.CertificateRequest {
	return &x509.CertificateRequest{
		Subject: pkix.Name{
			Country:            []string{"CN"},
			Organization:       []string{"Linux Foundation Edge"},
			OrganizationalUnit: []string{"BAETYL"},
			Locality:           []string{"Haidian District"},
			Province:           []string{"Beijing"},
			StreetAddress:      []string{"Baidu Campus"},
			PostalCode:         []string{"100093"},
			CommonName:         cn,
		},
		DNSNames:       alt.DNSNames,
		EmailAddresses: alt.Emails,
		IPAddresses:    alt.IPs,
		URIs:           alt.URIs,
	}
}
