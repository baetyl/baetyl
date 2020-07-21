package security

import (
	"encoding/base64"
	"errors"
	bh "github.com/timshannon/bolthold"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/baetyl/baetyl-go/v2/log"
	mc "github.com/baetyl/baetyl-go/v2/mock/pki"
	"github.com/baetyl/baetyl-go/v2/pki"
	"github.com/baetyl/baetyl-go/v2/pki/models"
	"github.com/baetyl/baetyl/config"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
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

func genPKIConf(t *testing.T) config.PKIConfig {
	tempDir, err := ioutil.TempDir("", "")
	assert.NoError(t, err)

	err = ioutil.WriteFile(path.Join(tempDir, "ca.crt"), []byte(caCrt), 777)
	assert.NoError(t, err)
	err = ioutil.WriteFile(path.Join(tempDir, "ca.key"), []byte(caKey), 777)
	assert.NoError(t, err)
	return config.PKIConfig{
		KeyFile: path.Join(tempDir, "ca.key"),
		CrtFile: path.Join(tempDir, "ca.crt"),
	}
}

func genDefaultPkiClient(t *testing.T) (*defaultPkiClient, *mc.MockStorage) {
	cfg := genPKIConf(t)
	ctl := gomock.NewController(t)
	mcSto := mc.NewMockStorage(ctl)
	cli, err := pki.NewPKIClient(cfg.KeyFile, cfg.CrtFile, mcSto)
	assert.NoError(t, err)
	assert.NotNil(t, cli)
	return &defaultPkiClient{
		cli: cli,
		sto: mcSto,
		cfg: cfg,
		log: log.With(log.Any("security", "pki")),
	}, mcSto
}

func genPkiMockCert() *models.Cert {
	return &models.Cert{
		CertId:     baetylSubCA,
		Content:    base64.StdEncoding.EncodeToString([]byte(caCrt)),
		PrivateKey: base64.StdEncoding.EncodeToString([]byte(caKey)),
	}
}

func Test_NewPKIImpl(t *testing.T) {
	// bad case
	_, err := newPKIImpl(config.SecurityConfig{}, nil)
	assert.Error(t, err)

	// good case
	cfg := config.SecurityConfig{
		Kind: "pki",
	}
	cfg.PKIConfig = genPKIConf(t)
	bhSto := genBolthold(t)
	res, err := newPKIImpl(cfg, bhSto)
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestDefaultPkiClient_GetCA(t *testing.T) {
	p, mcSto := genDefaultPkiClient(t)

	// good case
	cert := genPkiMockCert()
	mcSto.EXPECT().GetCert(baetylSubCA).Return(cert, nil).Times(1)

	res, err := p.GetCA()
	assert.NoError(t, err)
	assert.EqualValues(t, []byte(caCrt), res)

	//bad case
	mcSto.EXPECT().GetCert(baetylSubCA).Return(nil, bh.ErrNotFound).Times(1)
	_, err = p.GetCA()
	assert.Error(t, err, bh.ErrNotFound)
}

func TestDefaultPkiClient_IssueCertificate(t *testing.T) {
	p, mcSto := genDefaultPkiClient(t)

	cert := genPkiMockCert()
	// good case
	mcSto.EXPECT().GetCert(gomock.Any()).Return(cert, nil).Times(2)
	mcSto.EXPECT().CreateCert(gomock.Any()).Return(nil).Times(1)

	res, err := p.IssueCertificate("cn", AltNames{})
	assert.NoError(t, err)
	assert.EqualValues(t, caCrt, string(res.CertPEM))
}

func TestDefaultPkiClient_RevokeCertificate(t *testing.T) {
	p, mcSto := genDefaultPkiClient(t)

	// good case
	cert := genPkiMockCert()
	mcSto.EXPECT().DeleteCert(cert.CertId).Return(nil).Times(1)
	err := p.RevokeCertificate(cert.CertId)
	assert.NoError(t, err)

	// bad case
	mcSto.EXPECT().DeleteCert(cert.CertId).Return(os.ErrInvalid).Times(1)
	err = p.RevokeCertificate(cert.CertId)
	assert.Error(t, err)
}

func TestDefaultPkiClient_RotateCertificate(t *testing.T) {
	p, mcSto := genDefaultPkiClient(t)

	cert := genPkiMockCert()
	mcSto.EXPECT().GetCert(gomock.Any()).Return(cert, nil).Times(3)
	mcSto.EXPECT().DeleteCert(cert.CertId).Return(nil).Times(1)
	mcSto.EXPECT().CreateCert(gomock.Any()).Return(nil).Times(1)

	res, err := p.RotateCertificate(cert.CertId)
	assert.NoError(t, err)
	assert.EqualValues(t, caCrt, string(res.CertPEM))
}

func TestDefaultPkiClient_setSubCA(t *testing.T) {
	p, mcSto := genDefaultPkiClient(t)
	cert := genPkiMockCert()

	// good case 1:create new sub ca
	mcSto.EXPECT().GetCert(baetylSubCA).Return(nil, bh.ErrNotFound).Times(1)
	mcSto.EXPECT().CreateCert(*cert).Return(nil).Times(1)
	err := p.setSubCA()
	assert.NoError(t, err)

	// good case 2: no change
	mcSto.EXPECT().GetCert(baetylSubCA).Return(cert, nil).Times(1)
	err = p.setSubCA()
	assert.NoError(t, err)

	// good case 3: upgrade ca
	mcSto.EXPECT().GetCert(baetylSubCA).Return(&models.Cert{}, nil).Times(1)
	mcSto.EXPECT().UpdateCert(*cert).Return(nil).Times(1)
	err = p.setSubCA()
	assert.NoError(t, err)

	// bad case 1: get cert error
	mcSto.EXPECT().GetCert(baetylSubCA).Return(nil, errors.New("err")).Times(1)
	mcSto.EXPECT().CreateCert(*cert).Return(nil).Times(1)
	err = p.setSubCA()
	assert.Error(t, err)

	// bad case 2: read key file error
	p.cfg.KeyFile = ""
	err = p.setSubCA()
	assert.Error(t, err)

	// bad case 3: read crt file error
	p.cfg.CrtFile = ""
	err = p.setSubCA()
	assert.Error(t, err)
}
