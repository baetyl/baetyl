package initz

import (
	"fmt"
	"os"
	"path"
	"testing"

	v2plugin "github.com/baetyl/baetyl-go/v2/plugin"
	"github.com/baetyl/baetyl-go/v2/pubsub"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/stretchr/testify/assert"

	"github.com/baetyl/baetyl/v2/config"
)

const (
	certCa = `
-----BEGIN CERTIFICATE-----
MIICfjCCAiSgAwIBAgIIFja5hmwJjwAwCgYIKoZIzj0EAwIwgaUxCzAJBgNVBAYT
AkNOMRAwDgYDVQQIEwdCZWlqaW5nMRkwFwYDVQQHExBIYWlkaWFuIERpc3RyaWN0
MRUwEwYDVQQJEwxCYWlkdSBDYW1wdXMxDzANBgNVBBETBjEwMDA5MzEeMBwGA1UE
ChMVTGludXggRm91bmRhdGlvbiBFZGdlMQ8wDQYDVQQLEwZCQUVUWUwxEDAOBgNV
BAMTB3Jvb3QuY2EwIBcNMjAwOTIxMDY0NTA0WhgPMjA3MDA5MDkwNjQ1MDRaMIGl
MQswCQYDVQQGEwJDTjEQMA4GA1UECBMHQmVpamluZzEZMBcGA1UEBxMQSGFpZGlh
biBEaXN0cmljdDEVMBMGA1UECRMMQmFpZHUgQ2FtcHVzMQ8wDQYDVQQREwYxMDAw
OTMxHjAcBgNVBAoTFUxpbnV4IEZvdW5kYXRpb24gRWRnZTEPMA0GA1UECxMGQkFF
VFlMMRAwDgYDVQQDEwdyb290LmNhMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE
A3LIE7hL2K6c6HBOyXKySr756dDg65XswkX9rgexVgGr2JoOM2GEHTFix363Xgo9
GJNAUWh245RVtrnK8tcG+aM6MDgwDgYDVR0PAQH/BAQDAgGGMA8GA1UdEwEB/wQF
MAMBAf8wFQYDVR0RBA4wDIcEAAAAAIcEfwAAATAKBggqhkjOPQQDAgNIADBFAiAs
4bwazl2Ienqgf9J+QCRCJWX6huKx+Au+w/mvTOy+cwIhALgnhoBGS4PBd/xBEiOb
NxcsZX/LfkVYKvbwPbVBJjci
-----END CERTIFICATE-----
`
	certCrt = `
-----BEGIN CERTIFICATE-----
MIIChzCCAi6gAwIBAgIIFkHUUPfd0ggwCgYIKoZIzj0EAwIwgaUxCzAJBgNVBAYT
AkNOMRAwDgYDVQQIEwdCZWlqaW5nMRkwFwYDVQQHExBIYWlkaWFuIERpc3RyaWN0
MRUwEwYDVQQJEwxCYWlkdSBDYW1wdXMxDzANBgNVBBETBjEwMDA5MzEeMBwGA1UE
ChMVTGludXggRm91bmRhdGlvbiBFZGdlMQ8wDQYDVQQLEwZCQUVUWUwxEDAOBgNV
BAMTB3Jvb3QuY2EwHhcNMjAxMDI3MTA1OTQ2WhcNNDAxMDIyMTA1OTQ2WjCBrDEL
MAkGA1UEBhMCQ04xEDAOBgNVBAgTB0JlaWppbmcxGTAXBgNVBAcTEEhhaWRpYW4g
RGlzdHJpY3QxFTATBgNVBAkTDEJhaWR1IENhbXB1czEPMA0GA1UEERMGMTAwMDkz
MR4wHAYDVQQKExVMaW51eCBGb3VuZGF0aW9uIEVkZ2UxDzANBgNVBAsTBkJBRVRZ
TDEXMBUGA1UEAxMOZGVmYXVsdC4xMDI3MDEwWTATBgcqhkjOPQIBBggqhkjOPQMB
BwNCAATl+OSkL+Dptbt80PWuSiPMzUnkm5Si3xTFpNyG7nR6n7eaMJBnP0Jq4TNt
MyXdB8IJ6ohyWgaCrvZggN501zUxoz8wPTAOBgNVHQ8BAf8EBAMCBaAwHQYDVR0l
BBYwFAYIKwYBBQUHAwIGCCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwCgYIKoZIzj0E
AwIDRwAwRAIgF/7p7pT8HK1eKgGVUWCYezXEMAwIelwTcP7ottak2V8CIACMw6IO
GHGbGskkWEbfay6qymbOKQ3gCYB3L++rICS0
-----END CERTIFICATE-----
`
	certKey = `
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIKisyEJsO4SFqqfb9EjeaXPlj1BArNmwmZ/htCw/EcyPoAoGCCqGSM49
AwEHoUQDQgAE5fjkpC/g6bW7fND1rkojzM1J5JuUot8UxaTchu50ep+3mjCQZz9C
auEzbTMl3QfCCeqIcloGgq72YIDedNc1MQ==
-----END EC PRIVATE KEY-----
`
)

func TestNewInitialize(t *testing.T) {
	cfg := config.Config{}
	// with cert
	tmpDir := t.TempDir()

	certPath := path.Join(tmpDir, "cert.pem")
	err := os.WriteFile(certPath, []byte("cert"), 0755)
	assert.Nil(t, err)
	cfg.Node.Cert = certPath

	f, err := os.CreateTemp("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())
	cfg.Store.Path = f.Name()

	// bad case sync no plugin
	_, err = NewInitialize(cfg)
	assert.Error(t, err)

	// bad case engine ami err
	cfg.Plugin.Link = "link"
	v2plugin.RegisterFactory("link", func() (v2plugin.Plugin, error) {
		return &mockLink{}, nil
	})
	cfg.Plugin.Pubsub = "pubsub"
	v2plugin.RegisterFactory("pubsub", func() (v2plugin.Plugin, error) {
		res, err := pubsub.NewPubsub(1)
		assert.NoError(t, err)
		return res, nil
	})

	_, err = NewInitialize(cfg)
	assert.Error(t, err)

	// bad case err engine
	crt := path.Join(tmpDir, "crt.pem")
	err = os.WriteFile(crt, []byte(certCrt), 0755)
	assert.Nil(t, err)
	cfg.Node.Cert = crt
	ca := path.Join(tmpDir, "ca.pem")
	err = os.WriteFile(ca, []byte(certCa), 0755)
	assert.Nil(t, err)
	cfg.Node.CA = ca
	key := path.Join(tmpDir, "key.pem")
	err = os.WriteFile(key, []byte(certKey), 0755)
	assert.Nil(t, err)
	cfg.Node.Key = key

	_, err = NewInitialize(cfg)
	assert.Error(t, err)

	// bad case no cert
	cfg.Node.CA = ""
	cfg.Node.Cert = ""
	cfg.Node.Key = ""

	_, err = NewInitialize(cfg)
	assert.Error(t, err)
}

type mockLink struct{}

func (lk *mockLink) Close() error {
	return nil
}

func (lk *mockLink) Receive() (<-chan *specv1.Message, <-chan error) {
	return nil, nil
}

func (lk *mockLink) Request(msg *specv1.Message) (*specv1.Message, error) {
	return nil, nil
}

func (lk *mockLink) Send(msg *specv1.Message) error {
	return nil
}

func (lk *mockLink) State() *specv1.Message {
	return nil
}

func (lk *mockLink) IsAsyncSupported() bool {
	return false
}
