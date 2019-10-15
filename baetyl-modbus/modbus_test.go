package main

import (
	"encoding/binary"
	"fmt"
	"testing"
	"time"

	"github.com/goburrow/modbus"
	"github.com/stretchr/testify/assert"
)

func TestModbusUSB(t *testing.T) {
	// Modbus RTU/ASCII
	handler := modbus.NewRTUClientHandler("/dev/ttyUSB4")
	handler.BaudRate = 9600
	handler.DataBits = 8
	handler.Parity = "N"
	handler.StopBits = 1
	handler.SlaveId = 1
	handler.Timeout = 5 * time.Second
	// handler.RS485.Enabled = true

	err := handler.Connect()
	defer handler.Close()
	assert.NoError(t, err)

	client := modbus.NewClient(handler)
	results, err := client.ReadHoldingRegisters(0, 2)
	assert.NoError(t, err)
	assert.Equal(t, "R:", results)
	assert.Equal(t, "W", fmt.Sprintf("%d", binary.BigEndian.Uint16(results[0:2])))
	assert.Equal(t, "S", fmt.Sprintf("%d", binary.BigEndian.Uint16(results[2:])))
}

func TestModbusS(t *testing.T) {
	// Modbus RTU/ASCII
	handler := modbus.NewRTUClientHandler("/dev/ttyS0")
	handler.BaudRate = 9600
	handler.DataBits = 8
	handler.Parity = "N"
	handler.StopBits = 1
	handler.SlaveId = 1
	handler.Timeout = time.Second
	handler.RS485.Enabled = true
	// handler.RS485.RtsHighAfterSend = true
	// handler.RS485.RtsHighDuringSend = true
	// handler.RS485.RxDuringTx = true

	err := handler.Connect()
	defer handler.Close()
	assert.NoError(t, err)

	client := modbus.NewClient(handler)
	results, err := client.ReadHoldingRegisters(0, 2)
	assert.NoError(t, err)
	assert.Equal(t, "R:", results)
	assert.Equal(t, "W", fmt.Sprintf("%d", binary.BigEndian.Uint16(results[0:2])))
	assert.Equal(t, "S", fmt.Sprintf("%d", binary.BigEndian.Uint16(results[2:])))

}
