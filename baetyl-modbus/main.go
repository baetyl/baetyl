package main

import (
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/256dpi/gomqtt/packet"
	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
	"github.com/goburrow/modbus"
	"github.com/goburrow/serial"
)

// custom configuration of the timer module
type config struct {
	Modbus struct {
		// Address Device path (/dev/ttyS0)
		Address string `yaml:"address" json:"address" default:"/dev/ttyS0"`
		// Timeout Read (Write) timeout.
		Timeout time.Duration `yaml:"timeout" json:"timeout" default:"10s"`
		// IdleTimeout Idle timeout to close the connection
		IdleTimeout time.Duration `yaml:"idletimeout" json:"idletimeout" default:"1m"`

		//// RTU only
		// BaudRate (default 19200)
		BaudRate int `yaml:"baudrate" json:"baudrate" default:"19200"`
		// DataBits: 5, 6, 7 or 8 (default 8)
		DataBits int `yaml:"databits" json:"databits" default:"8"`
		// StopBits: 1 or 2 (default 1)
		StopBits int `yaml:"stopbits" json:"stopbits" default:"1"`
		// Parity: N - None, E - Even, O - Odd (default E)
		// (The use of no parity requires 2 stop bits.)
		Parity string `yaml:"parity" json:"parity" default:"E"`
		// RS485 Configuration related to RS485
		RS485 struct {
			// Enabled Enable RS485 support
			Enabled bool `yaml:"enabled" json:"enabled"`
			// DelayRtsBeforeSend Delay RTS prior to send
			DelayRtsBeforeSend time.Duration `yaml:"delay_rts_before_send" json:"delay_rts_before_send"`
			// DelayRtsAfterSend Delay RTS after send
			DelayRtsAfterSend time.Duration `yaml:"delay_rts_after_send" json:"delay_rts_after_send"`
			// RtsHighDuringSend Set RTS high during send
			RtsHighDuringSend bool `yaml:"rts_high_during_send" json:"rts_high_during_send"`
			// RtsHighAfterSend Set RTS high after send
			RtsHighAfterSend bool `yaml:"rts_high_after_send" json:"rts_high_after_send"`
			// RxDuringTx Rx during Tx
			RxDuringTx bool `yaml:"Rx_during_tx" json:"Rx_during_tx"`
		} `yaml:"rs485" json:"rs485"`
		// Slave
		Slave struct {
			// ID slave id
			ID byte `yaml:"id" json:"id"`
			// Function
			Function byte `yaml:"function" json:"function"`
			// Address
			Address uint16 `yaml:"address" json:"address"`
			// Quantity
			Quantity uint16 `yaml:"quantity" json:"quantity"`
			// Interval
			Interval time.Duration `yaml:"interval" json:"interval"`
		} `yaml:"slave" json:"slave"`
	} `yaml:"modbus" json:"modbus"`
	Publish struct {
		QOS   uint32 `yaml:"qos" json:"qos" validate:"min=0, max=1"`
		Topic string `yaml:"topic" json:"topic" default:"timer" validate:"nonzero"`
	} `yaml:"publish" json:"publish"`
}

func main() {
	// Running module in baetyl context
	baetyl.Run(func(ctx baetyl.Context) error {
		var cfg config
		// load custom config
		err := ctx.LoadConfig(&cfg)
		if err != nil {
			return err
		}
		// create a hub client
		cli, err := ctx.NewHubClient("", nil)
		if err != nil {
			return err
		}
		// start client to keep connection with hub
		cli.Start(nil)

		// modbus init
		var handler modbus.ClientHandler
		if strings.HasPrefix(cfg.Modbus.Address, "tcp://") {
			// Modbus TCP
			h := modbus.NewTCPClientHandler(cfg.Modbus.Address[6:])
			h.SlaveId = cfg.Modbus.Slave.ID
			h.Timeout = cfg.Modbus.Timeout
			h.IdleTimeout = cfg.Modbus.IdleTimeout
			err = h.Connect()
			if err != nil {
				return fmt.Errorf("Failed to connect: %s", err.Error())
			}
			defer h.Close()
			handler = h
		} else {
			// Modbus RTU/ASCII
			h := modbus.NewRTUClientHandler(cfg.Modbus.Address)
			h.BaudRate = cfg.Modbus.BaudRate
			h.DataBits = cfg.Modbus.DataBits
			h.Parity = cfg.Modbus.Parity
			h.StopBits = cfg.Modbus.StopBits
			h.SlaveId = cfg.Modbus.Slave.ID
			h.Timeout = cfg.Modbus.Timeout
			h.IdleTimeout = cfg.Modbus.IdleTimeout
			h.RS485 = serial.RS485Config{
				Enabled:            cfg.Modbus.RS485.Enabled,
				DelayRtsBeforeSend: cfg.Modbus.RS485.DelayRtsBeforeSend,
				DelayRtsAfterSend:  cfg.Modbus.RS485.DelayRtsAfterSend,
				RtsHighDuringSend:  cfg.Modbus.RS485.RtsHighDuringSend,
				RtsHighAfterSend:   cfg.Modbus.RS485.RtsHighAfterSend,
				RxDuringTx:         cfg.Modbus.RS485.RxDuringTx,
			}
			err = h.Connect()
			if err != nil {
				return fmt.Errorf("Failed to connect: %s", err.Error())
			}
			defer h.Close()
			handler = h
		}

		mb := modbus.NewClient(handler)

		// create a timer
		ticker := time.NewTicker(cfg.Modbus.Slave.Interval)
		defer ticker.Stop()
		var results []byte
		switch cfg.Modbus.Slave.Function {
		case 3:
			for {
				select {
				case t := <-ticker.C:
					results, err = mb.ReadHoldingRegisters(cfg.Modbus.Slave.Address, cfg.Modbus.Slave.Quantity)
					if err != nil {
						return fmt.Errorf("Failed to read: %s", err.Error())
					}

					pld := make([]byte, 4+cfg.Modbus.Slave.Quantity*2)
					binary.BigEndian.PutUint32(pld, uint32(t.Unix()))
					copy(pld[4:], results)

					pkt := packet.NewPublish()
					pkt.Message.Topic = cfg.Publish.Topic
					pkt.Message.QOS = packet.QOS(cfg.Publish.QOS)
					pkt.Message.Payload = pld
					// send a message to hub triggered by timer
					err = cli.Send(pkt)
					if err != nil {
						return fmt.Errorf("Failed to publish: %s", err.Error())
					}
				case <-ctx.WaitChan():
					// wait until service is stopped
					return nil
				}
			}
		default:
			return fmt.Errorf("Function code (%d) not supported", cfg.Modbus.Slave.Function)
		}
	})
}
