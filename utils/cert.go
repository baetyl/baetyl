package utils

import (
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
)

// GetSerialNumber gets serial number from pem
func GetSerialNumber(file string) (string, error) {
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}

	block, _ := pem.Decode(raw)
	x509Cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", err
	}
	return x509Cert.SerialNumber.String(), nil
}
