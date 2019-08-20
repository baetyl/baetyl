package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPortAvailable(t *testing.T) {
	got, err := GetAvailablePort("127.0.0.1")
	fmt.Println(got, err)
	got, err = GetAvailablePort("0.0.0.0")
	fmt.Println(got, err)
}

func TestCheckPortsInUse(t *testing.T) {
	ports := []string{"20000", "20001"}
	usedPorts, res := CheckPortsInUse(ports)
	assert.Equal(t, res, false)
	fmt.Println(usedPorts)
}
