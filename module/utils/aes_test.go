package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAesEncryptDecrypt(t *testing.T) {
	var tests = []struct {
		key, data string
	}{
		{
			"2b7e152356aed2a6abf7158809cf4f3c",
			"6bc1bee22e409f96e93d7e117393172a",
		},
		{
			"2b7e151628aed2a61256158809cf4f3c",
			"ae2d8a571e03ac9c9eb76fac45af8e51",
		},
		{
			"2b7e151628aed2a6abf7097509cf4f3c",
			"30c81c46a35ce411e5fbc1191a0a52ef",
		},
		{
			"2b7e151628aed2a6ab23678809cf4f3c",
			"f69f2445df4f9b17ad2b417be66c3710",
		},
		{
			"2b7e151628aed1245b23678809cf4f3c",
			"hello, tg boy, haha ??????,智能边缘",
		},
	}
	for _, test := range tests {
		assertEncryptDecrypt(t, []byte(test.key), []byte(test.data))
	}
}

func assertEncryptDecrypt(t *testing.T, key, data []byte) {
	encData, err := AesEncrypt(data, key)
	fmt.Println(encData)
	assert.Nil(t, err)
	decData, err := AesDecrypt(encData, key)
	assert.Nil(t, err)
	assertArrayEqual(t, data, decData)
}
