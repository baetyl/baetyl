package utils

import (
	"fmt"
	"testing"
)

func TestGetPortAvailable(t *testing.T) {
	got, err := GetPortAvailable("127.0.0.1")
	fmt.Println(got, err)
	got, err = GetPortAvailable("0.0.0.0")
	fmt.Println(got, err)
}
