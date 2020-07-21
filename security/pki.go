package security

import (
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"io/ioutil"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/pki"
	"github.com/baetyl/baetyl-go/v2/pki/models"
	"github.com/baetyl/baetyl/config"
	bh "github.com/timshannon/bolthold"
)

const (
	baetylSubCA = "baetylSubCA"
)

type defaultPkiClient struct {
	cli pki.PKI
	sto pki.Storage
	cfg config.PKIConfig
	log *log.Logger
}

func init() {
	Register(PKI, newPKIImpl)
}

func newPKIImpl(cfg config.SecurityConfig, bhSto *bh.Store) (Security, error) {
	sto := NewStorage(bhSto)
	cli, err := pki.NewPKIClient(cfg.PKIConfig.KeyFile, cfg.PKIConfig.CrtFile, sto)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defaultCli := &defaultPkiClient{
		cli: cli,
		sto: sto,
		cfg: cfg.PKIConfig,
		log: log.With(log.Any("security", cfg.Kind)),
	}
	err = defaultCli.setSubCA()
	if err != nil {
		return nil, errors.Trace(err)
	}
	return defaultCli, nil
}

func (p *defaultPkiClient) GetCA() ([]byte, error) {
	ca, err := p.cli.GetCert(baetylSubCA)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return ca, nil
}

func (p *defaultPkiClient) IssueCertificate(cn string, alt AltNames) (*PEMCredential, error) {
	priv, err := pki.GenCertPrivateKey(pki.DefaultDSA, pki.DefaultRSABits)
	if err != nil {
		return nil, errors.Trace(err)
	}

	keyPem, err := pki.EncodeCertPrivateKey(priv)
	if err != nil {
		return nil, errors.Trace(err)
	}

	csr, err := x509.CreateCertificateRequest(rand.Reader, genCsr(cn, alt), priv.Key)
	if err != nil {
		return nil, errors.Trace(err)
	}

	certId, err := p.cli.CreateSubCert(csr, baetylSubCA)
	if err != nil {
		return nil, errors.Trace(err)
	}
	certPem, err := p.cli.GetCert(certId)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &PEMCredential{
		CertPEM: certPem,
		KeyPEM:  keyPem,
		CertId:  certId,
	}, nil
}

func (p *defaultPkiClient) RevokeCertificate(certId string) error {
	err := p.cli.DeleteSubCert(certId)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (p *defaultPkiClient) RotateCertificate(certId string) (*PEMCredential, error) {
	cert, err := p.cli.GetCert(certId)
	if err != nil {
		return nil, errors.Trace(err)
	}
	err = p.RevokeCertificate(certId)
	if err != nil {
		return nil, errors.Trace(err)
	}
	certInfo, err := pki.ParseCertificates(cert)
	if err != nil {
		return nil, errors.Trace(err)
	}
	if len(certInfo) != 1 {
		return nil, errors.Trace(errors.Errorf("rotate certificate error (not exist)"))
	}

	alt := AltNames{
		DNSNames: certInfo[0].DNSNames,
		IPs:      certInfo[0].IPAddresses,
		Emails:   certInfo[0].EmailAddresses,
		URIs:     certInfo[0].URIs,
	}
	return p.IssueCertificate(certInfo[0].Subject.CommonName, alt)
}

func (p *defaultPkiClient) setSubCA() error {
	crt, err := ioutil.ReadFile(p.cfg.CrtFile)
	if err != nil {
		return errors.Trace(err)
	}
	priv, err := ioutil.ReadFile(p.cfg.KeyFile)
	if err != nil {
		return errors.Trace(err)
	}
	subCA := models.Cert{
		CertId:     baetylSubCA,
		Content:    base64.StdEncoding.EncodeToString(crt),
		PrivateKey: base64.StdEncoding.EncodeToString(priv),
	}

	res, err := p.sto.GetCert(subCA.CertId)
	if err != nil && err.Error() != bh.ErrNotFound.Error() {
		return errors.Trace(err)
	}
	if err != nil {
		return p.sto.CreateCert(subCA)
	}

	if res.Content != subCA.Content || res.PrivateKey != subCA.PrivateKey {
		// TODO: when the sub-root certificate is updated, the certificates of all modules need to be re-issued
		return p.sto.UpdateCert(subCA)
	}
	return nil
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
