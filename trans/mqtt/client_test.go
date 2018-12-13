package mqtt_test

import (
	"testing"
	"time"

	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/transport/flow"
	"github.com/baidu/openedge/trans/mqtt"
	"github.com/stretchr/testify/assert"
)

func TestClientConnectErrorMissingAddress(t *testing.T) {
	c, err := mqtt.NewClient(mqtt.ClientConfig{}, nil)
	assert.EqualError(t, err, "parse : empty url")
	assert.Nil(t, c)
}

func TestClientConnectErrorWrongPort(t *testing.T) {
	cc := newConfig(t, "1234567")
	c, err := mqtt.NewClient(cc, assertNoErrorCallback(t))
	assert.EqualError(t, err, "dial tcp: address 1234567: invalid port")
	assert.Nil(t, c)
}

func TestClientConnect(t *testing.T) {
	broker := flow.New().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(t, port)
	c, err := mqtt.NewClient(cc, assertNoErrorCallback(t))
	assert.NoError(t, err)
	assert.NotNil(t, c)

	err = c.Close()
	assert.NoError(t, err)

	safeReceive(done)
}

func TestClientConnectCustomDialer(t *testing.T) {
	broker := flow.New().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(t, port)
	c, err := mqtt.NewClient(cc, assertNoErrorCallback(t))
	assert.NoError(t, err)
	assert.NotNil(t, c)

	err = c.Close()
	assert.NoError(t, err)

	safeReceive(done)
}

func TestClientConnectWithCredentials(t *testing.T) {
	connect := connectPacket()
	connect.Username = "test"
	connect.Password = "test"

	connack := connackPacket()
	connack.ReturnCode = packet.BadUsernameOrPassword

	broker := flow.New().Debug().
		Receive(connect).
		Send(connack).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(t, port)
	cc.Username = "test"
	cc.Password = "test"
	cb := func(p packet.Generic, err error) {
		assert.EqualError(t, err, "connection refused: bad user name or password")
	}
	c, err := mqtt.NewClient(cc, cb)
	assert.EqualError(t, err, "connection refused: bad user name or password")
	assert.Nil(t, c)

	safeReceive(done)
}

func TestClientConnectionDenied(t *testing.T) {
	connack := connackPacket()
	connack.ReturnCode = packet.NotAuthorized

	broker := flow.New().Debug().
		Receive(connectPacket()).
		Send(connack).
		Close()

	done, port := fakeBroker(t, broker)

	cc := newConfig(t, port)
	cb := func(p packet.Generic, err error) {
		assert.EqualError(t, err, "connection refused: not authorized")
	}
	c, err := mqtt.NewClient(cc, cb)
	assert.Nil(t, c)
	assert.EqualError(t, err, "connection refused: not authorized")

	safeReceive(done)
}

func TestClientExpectedConnack(t *testing.T) {
	broker := flow.New().Debug().
		Receive(connectPacket()).
		Send(packet.NewPingresp()).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(t, port)
	cb := func(p packet.Generic, err error) {
		assert.EqualError(t, err, "client expected connack")
	}
	c, err := mqtt.NewClient(cc, cb)
	assert.Nil(t, c)
	assert.EqualError(t, err, "client expected connack")

	safeReceive(done)
}

func TestClientNotExpectedConnack(t *testing.T) {
	broker := flow.New().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Send(connackPacket()).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(t, port)
	cb := func(p packet.Generic, err error) {
		assert.EqualError(t, err, "client already connecting")
	}
	c, err := mqtt.NewClient(cc, cb)
	assert.NoError(t, err)

	safeReceive(done)

	err = c.Close()
	assert.EqualError(t, err, "client already connecting")
}

func TestClientKeepAlive(t *testing.T) {
	connect := connectPacket()
	connect.KeepAlive = 0

	pingreq := packet.NewPingreq()
	pingresp := packet.NewPingresp()

	broker := flow.New().Debug().
		Receive(connect).
		Send(connackPacket()).
		Receive(pingreq).
		Send(pingresp).
		Receive(pingreq).
		Send(pingresp).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(t, port)
	cc.KeepAlive = time.Millisecond * 100
	c, err := mqtt.NewClient(cc, assertNoErrorCallback(t))
	assert.NoError(t, err)

	<-time.After(250 * time.Millisecond)

	err = c.Close()
	assert.NoError(t, err)

	safeReceive(done)
}

func TestClientKeepAliveTimeout(t *testing.T) {
	connect := connectPacket()
	connect.KeepAlive = 0

	pingreq := packet.NewPingreq()

	broker := flow.New().Debug().
		Receive(connect).
		Send(connackPacket()).
		Receive(pingreq).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(t, port)
	cc.KeepAlive = time.Millisecond * 5
	cb := func(p packet.Generic, err error) {
		assert.EqualError(t, err, "client missing pong")
	}
	c, err := mqtt.NewClient(cc, cb)
	assert.NoError(t, err)

	safeReceive(done)

	err = c.Close()
	assert.EqualError(t, err, "client missing pong")
}

func TestClientPublishSubscribeQOS0(t *testing.T) {
	subscribe := packet.NewSubscribe()
	subscribe.Subscriptions = []packet.Subscription{{Topic: "test"}}
	subscribe.ID = 1

	suback := packet.NewSuback()
	suback.ReturnCodes = []packet.QOS{0}
	suback.ID = 1

	publish := packet.NewPublish()
	publish.Message.Topic = "test"
	publish.Message.Payload = []byte("test")

	broker := flow.New().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(subscribe).
		Send(suback).
		Receive(publish).
		Send(publish).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	wait := make(chan struct{})

	callback := func(pkt packet.Generic, err error) {
		assert.NoError(t, err)
		p, ok := pkt.(*packet.Publish)
		assert.True(t, ok)
		assert.Equal(t, "test", p.Message.Topic)
		assert.Equal(t, []byte("test"), p.Message.Payload)
		assert.Equal(t, packet.QOS(0), p.Message.QOS)
		assert.False(t, p.Message.Retain)
		close(wait)
	}
	cc := newConfig(t, port)
	cc.Subscriptions = []mqtt.Subscription{{Topic: "test"}}
	c, err := mqtt.NewClient(cc, callback)
	assert.NoError(t, err)

	err = c.Send(publish)
	assert.NoError(t, err)

	safeReceive(wait)

	err = c.Close()
	assert.NoError(t, err)

	safeReceive(done)
}

func TestClientPublishSubscribeQOS1(t *testing.T) {
	subscribe := packet.NewSubscribe()
	subscribe.Subscriptions = []packet.Subscription{{Topic: "test", QOS: 1}}
	subscribe.ID = 1

	suback := packet.NewSuback()
	suback.ReturnCodes = []packet.QOS{1}
	suback.ID = 1

	publish := packet.NewPublish()
	publish.Message.Topic = "test"
	publish.Message.Payload = []byte("test")
	publish.Message.QOS = 1
	publish.ID = 2

	puback := packet.NewPuback()
	puback.ID = 2

	broker := flow.New().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(subscribe).
		Send(suback).
		Receive(publish).
		Send(puback).
		Send(publish).
		Receive(puback).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	wait := make(chan struct{})

	callback := func(pkt packet.Generic, err error) {
		assert.NoError(t, err)
		switch p := pkt.(type) {
		case *packet.Publish:
			assert.Equal(t, "test", p.Message.Topic)
			assert.Equal(t, []byte("test"), p.Message.Payload)
			assert.Equal(t, packet.QOS(1), p.Message.QOS)
			assert.False(t, p.Message.Retain)
			close(wait)
		case *packet.Puback:
			assert.Equal(t, packet.ID(2), p.ID)
		}
	}
	cc := newConfig(t, port)
	cc.Subscriptions = []mqtt.Subscription{{Topic: "test", QOS: 1}}
	c, err := mqtt.NewClient(cc, callback)
	assert.NoError(t, err)

	err = c.Send(publish)
	assert.NoError(t, err)

	safeReceive(wait)

	err = c.Send(puback)
	assert.NoError(t, err)

	err = c.Close()
	assert.NoError(t, err)

	safeReceive(done)
}

func TestClientUnexpectedClose(t *testing.T) {
	broker := flow.New().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Close()

	done, port := fakeBroker(t, broker)

	cc := newConfig(t, port)
	cb := func(p packet.Generic, err error) {
		assert.EqualError(t, err, "client not connected")
	}
	c, err := mqtt.NewClient(cc, cb)
	assert.NoError(t, err)

	safeReceive(done)
	time.Sleep(time.Millisecond * 100)

	err = c.Close()
	assert.EqualError(t, err, "client not connected")
}

func TestClientConnackFutureCancellation(t *testing.T) {
	broker := flow.New().Debug().
		Receive(connectPacket()).
		Close()

	done, port := fakeBroker(t, broker)

	cc := newConfig(t, port)
	cb := func(p packet.Generic, err error) {
		assert.EqualError(t, err, "client not connected")
	}
	c, err := mqtt.NewClient(cc, cb)
	assert.Nil(t, c)
	assert.EqualError(t, err, "client not connected")

	safeReceive(done)
}

func TestClientConnackFutureTimeout(t *testing.T) {
	broker := flow.New().Debug().
		Receive(connectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(t, port)
	cc.Timeout = time.Millisecond * 50
	cb := func(p packet.Generic, err error) {
		assert.EqualError(t, err, "future timeout")
	}
	c, err := mqtt.NewClient(cc, cb)
	assert.Nil(t, c)
	assert.EqualError(t, err, "future timeout")

	safeReceive(done)
}
