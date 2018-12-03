package utils

import (
	"bytes"
	"math/rand"
	"strings"
)

// random string char fields
const randomStrChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_"

// GenRandomStr generate specified length random string
func GenRandomStr(length int) string {
	var buf bytes.Buffer
	strArray := strings.Split(randomStrChars, "")
	for i := 0; i < length; i++ {
		buf.WriteString(strArray[rand.Intn(len(strArray))])
	}
	return buf.String()
}
