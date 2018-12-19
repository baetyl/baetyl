package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	pubKey = []byte(`
-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDZfKVxz7eoITkGj8GTZuuGyx1l
CjYbyamsA6UFwLtV4gDttaCcumChO8eIrGEEuThhqC2u7WFKjFazmP7DYoPyheUx
DjkUn1CJxaoSTkSlghN4XJ22XAqqrpsjloO3j6UHmsQokHpdrzJv2B/o+ojjkcH5
5IC1aeGBYM4XDb2o8wIDAQAB
-----END PUBLIC KEY-----`)
	priKey = []byte(`
-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQDZfKVxz7eoITkGj8GTZuuGyx1lCjYbyamsA6UFwLtV4gDttaCc
umChO8eIrGEEuThhqC2u7WFKjFazmP7DYoPyheUxDjkUn1CJxaoSTkSlghN4XJ22
XAqqrpsjloO3j6UHmsQokHpdrzJv2B/o+ojjkcH55IC1aeGBYM4XDb2o8wIDAQAB
AoGALgqXM7rXlH5EBkGUp1HYdpa1SFibD9LnWoUDAG7Gue24aJpUwBksr7VqDmL/
vvI/H11tHmUefZusFyVCebZ3XBI4AjMy3KJv2w4s+xjhN/2C7YgT4oMyjq7uhh8l
2n9Jw6KCsGnQVDgx8xSvbsb654U8xuViG2/Ugnyb3NFtWTECQQD4MQO27sDGljwB
nl3tewn5d1Ej3PBqtaQTk0Lji9cAdyJi3QUvl7yvidItFDboa+Wuyti3R0Tv+8zE
EBv2dZ0FAkEA4FRUsd52JJepI3DVcjyxlqc974wQZOTzrvkrsXSnmJW5l+IQu54P
vJCa43+oD/EaqzLWS99qnBIrIDkaB3vPlwJALWNBS6Xz6R02UhF1GeXjWBTC6O0R
pmIbZF0M4XIEWphu2GeU+DQmlG9+2TGWLQD2WvXLlhDZgY2pz70mb/boRQJACoPp
dGz5HL3/L6oaV0CBEo7EWHY4ToJs6cbERY0yTfS2vmfaYPEHy877c66IMjcbCOtZ
IDVYyfgQDXKfxboIAQJBAKER+qu/jjGKNx+3meXvqaRpYrutFDFZtXfHFjbYtrnk
fbdaHS+VNvZ2hkriZfdrJrLCTKiXVCPN2QHpBCPl0NU=
-----END RSA PRIVATE KEY-----`)
)

func TestRSAEncryptDecrypt(t *testing.T) {
	var tests = [][]byte{
		[]byte("hello, tg boy, haha ??????,智能边缘"),
		[]byte{0xff, 0xff, 0xff, 0x00, 0x00},
		[]byte{0x00, 0x00, 0x00, 0x00, 0x00},
		[]byte{0xff, 0xff, 0xff, 0xff, 0xff},
		[]byte{0xff},
		[]byte{0x00},
		[]byte{},
	}
	for _, data := range tests {
		assertPrivateEncryptPublicDecrypt(t, data)
		assertPublicEncryptPrivateDecrypt(t, data)
	}
}

func assertPrivateEncryptPublicDecrypt(t *testing.T, data []byte) {
	EncData, err := RsaPrivateEncrypt(data, priKey)
	assert.Nil(t, err)
	DecData, err := RsaPublicDecrypt(EncData, pubKey)
	assert.Nil(t, err)
	assertArrayEqual(t, data, DecData)
}

func assertPublicEncryptPrivateDecrypt(t *testing.T, data []byte) {
	EncData, err := RsaPublicEncrypt(data, pubKey)
	assert.Nil(t, err)
	DecData, err := RsaPrivateDecrypt(EncData, priKey)
	assert.Nil(t, err)
	assertArrayEqual(t, data, DecData)
}

func assertArrayEqual(t *testing.T, a []byte, b []byte) {
	if len(a) != len(b) {
		t.Fatal("array not equal!")
		return
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			t.Fatalf("array not equal at position %d, %d != %d", i, a[i], b[i])
			return
		}
	}
}
