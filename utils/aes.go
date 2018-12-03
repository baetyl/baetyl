package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"math/rand"

	"github.com/juju/errors"
)

// AES aes
const (
	AesIvLen      = 16
	AesKeyLen     = 32
	AesKeyCharSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

// AesEncrypt encrypts data using the specified key
func AesEncrypt(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Trace(err)
	}
	plaintext = paddingPKCS7(plaintext, block.BlockSize())
	ciphertext := make([]byte, len(plaintext))
	iv := make([]byte, AesIvLen)

	blockModel := cipher.NewCBCEncrypter(block, iv)
	blockModel.CryptBlocks(ciphertext, plaintext)
	return ciphertext, nil
}

// AesDecrypt decrypts data using the specified key
func AesDecrypt(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Trace(err)
	}

	iv := make([]byte, AesIvLen)
	plaintext := make([]byte, len(ciphertext))

	blockModel := cipher.NewCBCDecrypter(block, iv)
	blockModel.CryptBlocks(plaintext, ciphertext)
	plaintext = unpaddingPKCS7(plaintext, block.BlockSize())
	return plaintext, nil
}

// NewAesKey news AES key
func NewAesKey() []byte {
	key := make([]byte, AesKeyLen)
	for i := range key {
		key[i] = AesKeyCharSet[rand.Intn(len(AesKeyCharSet))]
	}
	return key
}

func paddingPKCS7(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func unpaddingPKCS7(plaintext []byte, blockSize int) []byte {
	length := len(plaintext)
	unpadding := int(plaintext[length-1])
	return plaintext[:(length - unpadding)]
}
