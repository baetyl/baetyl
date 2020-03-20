package baetyl

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const sysCertFile = "system.cert"
const sysCertName = "api"

type certificate struct {
	caraw []byte
	ca    *x509.Certificate
	cakey *rsa.PrivateKey
	cert  tls.Certificate
	pool  *x509.CertPool
}

func (rt *runtime) setupSysCert() error {
	err := rt.loadSysCert()
	if err != nil {
		err = rt.genSysCert()
	}
	return err
}

func (rt *runtime) loadSysCert() error {
	data, err := ioutil.ReadFile(filepath.Join(rt.cfg.DataPath, sysCertFile))
	if err != nil {
		return err
	}
	var block *pem.Block
	block, data = pem.Decode(data)
	if block == nil || block.Type != "CERTIFICATE" {
		return errors.New("first one should be a certificate")
	}
	rt.cert.caraw = make([]byte, len(block.Bytes))
	copy(rt.cert.caraw, block.Bytes)
	rt.cert.ca, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		return err
	}
	block, data = pem.Decode(data)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return errors.New("second one should be a certificate")
	}
	rt.cert.cakey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return err
	}
	block, data = pem.Decode(data)
	if block == nil {
		return errors.New("null pem data")
	}
	certPEM := pem.EncodeToMemory(block)
	block, data = pem.Decode(data)
	if block == nil {
		return errors.New("null pem data")
	}
	keyPEM := pem.EncodeToMemory(block)
	rt.cert.cert, err = tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return err
	}
	rt.cert.pool = x509.NewCertPool()
	rt.cert.pool.AddCert(rt.cert.ca)
	return nil
}

func (rt *runtime) genSysCert() error {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2020),
		Subject: pkix.Name{
			Organization:  []string{"baetyl.io"},
			Country:       []string{"CN"},
			Province:      []string{""},
			Locality:      []string{"Beijing"},
			StreetAddress: []string{"Baidu Campus"},
			PostalCode:    []string{"100093"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	var err error
	rt.cert.cakey, err = rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}
	rt.cert.caraw, err = x509.CreateCertificate(
		rand.Reader,
		ca, ca,
		&rt.cert.cakey.PublicKey, rt.cert.cakey,
	)
	if err != nil {
		return err
	}
	rt.cert.ca, err = x509.ParseCertificate(rt.cert.caraw)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	err = pem.Encode(&buf, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: rt.cert.caraw,
	})
	if err != nil {
		return err
	}
	err = pem.Encode(&buf, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(rt.cert.cakey),
	})
	if err != nil {
		return err
	}
	rt.cert.pool = x509.NewCertPool()
	rt.cert.pool.AddCert(rt.cert.ca)
	certBytes, certkey, err := rt.signCert(sysCertName)
	if err != nil {
		return err
	}
	err = pem.Encode(&buf, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err != nil {
		return err
	}
	err = pem.Encode(&buf, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certkey),
	})
	if err != nil {
		return err
	}
	rt.cert.cert.Certificate = append(rt.cert.cert.Certificate, certBytes)
	rt.cert.cert.PrivateKey = certkey
	return ioutil.WriteFile(filepath.Join(rt.cfg.DataPath, sysCertFile), buf.Bytes(), 0600)
}

func (rt *runtime) genServiceCert(file, name string) error {
	fi, err := os.Stat(file)
	if err == nil && fi.Mode().IsRegular() {
		return nil
	}
	certBytes, certkey, err := rt.signCert(name)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	err = pem.Encode(&buf, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err != nil {
		return err
	}
	err = pem.Encode(&buf, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certkey),
	})
	if err != nil {
		return err
	}
	err = pem.Encode(&buf, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: rt.cert.caraw,
	})
	if err != nil {
		return err
	}
	return ioutil.WriteFile(file, buf.Bytes(), 0600)
}

func (rt *runtime) signCert(name string) ([]byte, *rsa.PrivateKey, error) {
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			Organization:  []string{"baetyl.io"},
			Country:       []string{"CN"},
			Province:      []string{""},
			Locality:      []string{"Beijing"},
			StreetAddress: []string{"Baidu Campus"},
			PostalCode:    []string{"100093"},
			CommonName:    fmt.Sprintf("%s.baetyl.local", name),
		},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(10, 0, 0),
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}
	certkey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}
	certBytes, err := x509.CreateCertificate(
		rand.Reader,
		cert, rt.cert.ca,
		&certkey.PublicKey, rt.cert.cakey,
	)
	if err != nil {
		return nil, nil, err
	}
	return certBytes, certkey, nil
}

type certChain struct {
	cert tls.Certificate
	pool *x509.CertPool
}

func loadCertChain(file string) (certChain, error) {
	fail := func(err error) (certChain, error) { return certChain{}, err }

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return fail(err)
	}
	var block *pem.Block
	var certRaw []byte
	block, data = pem.Decode(data)
	if block == nil || block.Type != "CERTIFICATE" {
		return fail(errors.New("first one should be a certificate"))
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fail(err)
	}
	certRaw = block.Bytes
	block, data = pem.Decode(data)
	if block == nil || !strings.HasSuffix(block.Type, "PRIVATE KEY") {
		return fail(errors.New("second one should be a certificate"))
	}
	var key crypto.PrivateKey
	key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		key, err = x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			key, err = x509.ParseECPrivateKey(block.Bytes)
			if err != nil {
				return fail(err)
			}
		}
		switch key.(type) {
		case *rsa.PrivateKey, *ecdsa.PrivateKey, ed25519.PrivateKey:
			break
		default:
			return fail(errors.New("tls: found unknown private key type"))
		}
	}
	switch pub := cert.PublicKey.(type) {
	case *rsa.PublicKey:
		priv, ok := key.(*rsa.PrivateKey)
		if !ok {
			return fail(errors.New("tls: private key type does not match public key type"))
		}
		if pub.N.Cmp(priv.N) != 0 {
			return fail(errors.New("tls: private key does not match public key"))
		}
	case *ecdsa.PublicKey:
		priv, ok := key.(*ecdsa.PrivateKey)
		if !ok {
			return fail(errors.New("tls: private key type does not match public key type"))
		}
		if pub.X.Cmp(priv.X) != 0 || pub.Y.Cmp(priv.Y) != 0 {
			return fail(errors.New("tls: private key does not match public key"))
		}
	case ed25519.PublicKey:
		priv, ok := key.(ed25519.PrivateKey)
		if !ok {
			return fail(errors.New("tls: private key type does not match public key type"))
		}
		if !bytes.Equal(priv.Public().(ed25519.PublicKey), pub) {
			return fail(errors.New("tls: private key does not match public key"))
		}
	default:
		return fail(errors.New("tls: unknown public key algorithm"))
	}
	pool := x509.NewCertPool()
	for {
		block, data = pem.Decode(data)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" {
			continue
		}
		x509cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return fail(err)
		}
		pool.AddCert(x509cert)
	}
	return certChain{
		cert: tls.Certificate{
			Certificate: [][]byte{certRaw},
			PrivateKey:  key,
			Leaf:        cert,
		},
		pool: pool,
	}, nil
}
