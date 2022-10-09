package security

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/pki"
	"github.com/stretchr/testify/assert"
	bh "github.com/timshannon/bolthold"

	"github.com/baetyl/baetyl/v2/config"
	"github.com/baetyl/baetyl/v2/store"
	"github.com/baetyl/baetyl/v2/sync"
)

const (
	caKey = `
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIMQvliv2uMC+adF4q1SzJvjO/JFrrkbR5W1MeCvxtDIqoAoGCCqGSM49
AwEHoUQDQgAE/gdSFh3a6vA33+WMbBUAKB02L4bIl7hxN+436mC1zByzkmO7vaFm
xsymi2SRxVkLqMivlpWpMfbp5o21aMzpOw==
-----END EC PRIVATE KEY-----
`
	caCrt = `
-----BEGIN CERTIFICATE-----
MIICYzCCAgigAwIBAgIDAYaiMAoGCCqGSM49BAMCMIGlMQswCQYDVQQGEwJDTjEQ
MA4GA1UECBMHQmVpamluZzEZMBcGA1UEBxMQSGFpZGlhbiBEaXN0cmljdDEVMBMG
A1UECRMMQmFpZHUgQ2FtcHVzMQ8wDQYDVQQREwYxMDAwOTMxHjAcBgNVBAoTFUxp
bnV4IEZvdW5kYXRpb24gRWRnZTEPMA0GA1UECxMGQkFFVFlMMRAwDgYDVQQDEwdy
b290LmNhMCAXDTIwMDcyMTAzNTk1N1oYDzIwNzAwNzIxMDM1OTU3WjCBpTELMAkG
A1UEBhMCQ04xEDAOBgNVBAgTB0JlaWppbmcxGTAXBgNVBAcTEEhhaWRpYW4gRGlz
dHJpY3QxFTATBgNVBAkTDEJhaWR1IENhbXB1czEPMA0GA1UEERMGMTAwMDkzMR4w
HAYDVQQKExVMaW51eCBGb3VuZGF0aW9uIEVkZ2UxDzANBgNVBAsTBkJBRVRZTDEQ
MA4GA1UEAxMHcm9vdC5jYTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABP4HUhYd
2urwN9/ljGwVACgdNi+GyJe4cTfuN+pgtcwcs5Jju72hZsbMpotkkcVZC6jIr5aV
qTH26eaNtWjM6TujIzAhMA4GA1UdDwEB/wQEAwIBhjAPBgNVHRMBAf8EBTADAQH/
MAoGCCqGSM49BAMCA0kAMEYCIQDaTuoQ9CMNNRFKT5vFI8cvz1oZ4xQtkqtvk/p3
/xYCHQIhAMr6KcYHC9V0NDT9YhdQDN9J8z58QaUuOHoD8UV9VOG7
-----END CERTIFICATE-----
`
)

func genBolthold(t *testing.T) *bh.Store {
	f, err := os.CreateTemp("", "")
	assert.NoError(t, err)
	assert.NotNil(t, f)

	sto, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, sto)
	return sto
}

func genDefaultPkiClient(t *testing.T) *pkiClient {
	sto := genBolthold(t)
	cli, err := pki.NewPKIClient()
	assert.NoError(t, err)
	assert.NotNil(t, cli)
	return &pkiClient{
		cli: cli,
		sto: sto,
		cfg: config.SecurityConfig{
			Kind: "pki",
			PKIConfig: config.PKIConfig{
				SubDuration:  5 * 365 * 24 * time.Hour,
				RootDuration: 10 * 365 * 24 * time.Hour,
			},
		},
		log: log.With(log.Any("security", "pki")),
	}
}

func Test_NewPKI(t *testing.T) {
	// good case
	sto := genBolthold(t)
	_, err := NewPKI(config.SecurityConfig{
		Kind: "pki",
		PKIConfig: config.PKIConfig{
			SubDuration:  5 * 365 * 24 * time.Hour,
			RootDuration: 10 * 365 * 24 * time.Hour,
		},
	}, sto)
	assert.NoError(t, err)
}

func TestDefaultPkiClient_GetCA(t *testing.T) {
	p := genDefaultPkiClient(t)

	// bad case
	res, err := p.GetCA()
	assert.Error(t, err, bh.ErrNotFound)
	assert.Nil(t, res)

	// good case
	cn := fmt.Sprintf("%s.%s", os.Getenv(sync.EnvKeyNodeNamespace), os.Getenv(sync.EnvKeyNodeName))
	key := genStoKey(cn)
	cert := pki.CertPem{
		Crt: []byte(caCrt),
		Key: []byte(caKey),
	}
	err = p.putCert(key, cert)
	assert.NoError(t, err)

	res, err = p.GetCA()
	assert.NoError(t, err)
	assert.EqualValues(t, cert.Crt, res)
}

func TestDefaultPkiClient_IssueCertificate(t *testing.T) {
	p := genDefaultPkiClient(t)
	// bad case
	res, err := p.IssueCertificate("cn", AltNames{})
	assert.Error(t, err, bh.ErrNotFound)
	assert.Nil(t, res)

	// good case
	cn := fmt.Sprintf("%s.%s", os.Getenv(sync.EnvKeyNodeNamespace), os.Getenv(sync.EnvKeyNodeName))
	key := genStoKey(cn)
	cert := pki.CertPem{
		Crt: []byte(caCrt),
		Key: []byte(caKey),
	}
	err = p.putCert(key, cert)
	assert.NoError(t, err)

	res, err = p.IssueCertificate("cn", AltNames{})
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestDefaultPkiClient_RevokeCertificate(t *testing.T) {
	p := genDefaultPkiClient(t)
	// bad case
	err := p.RevokeCertificate("cn")
	assert.Error(t, err, bh.ErrNotFound)

	// good case
	key := genStoKey("cn")
	cert := pki.CertPem{
		Crt: []byte(caCrt),
		Key: []byte(caKey),
	}
	err = p.putCert(key, cert)
	assert.NoError(t, err)

	err = p.RevokeCertificate("cn")
	assert.NoError(t, err)
}

func TestDefaultPkiClient_RotateCertificate(t *testing.T) {
	p := genDefaultPkiClient(t)
	// bad case
	res, err := p.RotateCertificate("cn")
	assert.Error(t, err, bh.ErrNotFound)
	assert.Nil(t, res)

	// good case
	rootKey := genStoKey(fmt.Sprintf("%s.%s", os.Getenv(sync.EnvKeyNodeNamespace), os.Getenv(sync.EnvKeyNodeName)))
	key := genStoKey("cn")
	cert := pki.CertPem{
		Crt: []byte(caCrt),
		Key: []byte(caKey),
	}
	err = p.putCert(rootKey, cert) // root ca
	assert.NoError(t, err)
	err = p.putCert(key, cert) // old cert
	assert.NoError(t, err)

	res, err = p.RotateCertificate("cn")
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestDefaultPkiClient_putCert(t *testing.T) {
	p := genDefaultPkiClient(t)
	key := genStoKey("cn")
	cert := pki.CertPem{
		Crt: []byte(caCrt),
		Key: []byte(caKey),
	}

	// good case 0
	err := p.putCert(key, cert)
	assert.NoError(t, err)

	// good case 1
	err = p.putCert(key, cert)
	assert.NoError(t, err)
}
