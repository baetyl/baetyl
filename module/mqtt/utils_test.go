package mqtt_test

import (
	"net"
	"testing"
	"time"

	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/transport"
	"github.com/256dpi/gomqtt/transport/flow"
	"github.com/baidu/openedge/module/config"
	"github.com/baidu/openedge/module/mqtt"
	"github.com/creasty/defaults"
	"github.com/stretchr/testify/assert"
)

func safeReceive(ch chan struct{}) {
	select {
	case <-time.After(1 * time.Minute):
		panic("nothing received")
	case <-ch:
	}
}

func newConfig(t *testing.T, port string) (c config.MQTTClient) {
	c.CleanSession = true
	c.Address = "tcp://localhost:" + port
	defaults.Set(&c)
	return
}

func assertNoErrorCallback(t *testing.T) (h mqtt.Handler) {
	h.ProcessError = func(err error) {
		assert.NoError(t, err)
		assert.FailNow(t, "ProcessError should not have been called")
	}
	return
}

func fakeBroker(t *testing.T, testFlows ...*flow.Flow) (chan struct{}, string) {
	done := make(chan struct{})

	server, err := transport.Launch("tcp://localhost:0")
	assert.NoError(t, err)

	go func() {
		for _, flow := range testFlows {
			conn, err := server.Accept()
			assert.NoError(t, err)

			err = flow.Test(conn)
			assert.NoError(t, err)
		}

		err = server.Close()
		assert.NoError(t, err)

		close(done)
	}()

	_, port, _ := net.SplitHostPort(server.Addr().String())

	return done, port
}

func connectPacket() *packet.Connect {
	pkt := packet.NewConnect()
	pkt.CleanSession = true
	pkt.KeepAlive = 60
	return pkt
}

func connackPacket() *packet.Connack {
	pkt := packet.NewConnack()
	pkt.ReturnCode = packet.ConnectionAccepted
	pkt.SessionPresent = false
	return pkt
}

func disconnectPacket() *packet.Disconnect {
	return packet.NewDisconnect()
}
