package utils

import (
	"fmt"
	"testing"
)

func TestGetPortAvailable(t *testing.T) {
	got, err := GetAvailablePort("127.0.0.1")
	fmt.Println(got, err)
	got, err = GetAvailablePort("0.0.0.0")
	fmt.Println(got, err)
}
