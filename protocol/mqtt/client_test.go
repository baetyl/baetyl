package mqtt

import (
	"testing"
	"time"

	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/transport/flow"
	"github.com/stretchr/testify/assert"
)

type mackHandler struct {
	t                      *testing.T
	expectedError          string
	expectedProcessPublish func(*packet.Publish) error
	expectedProcessPuback  func(*packet.Puback) error
}

func (h *mackHandler) ProcessPublish(pkt *packet.Publish) error {
	if h.expectedProcessPublish != nil {
		return h.expectedProcessPublish(pkt)
	}
	return nil
}

func (h *mackHandler) ProcessPuback(pkt *packet.Puback) error {
	if h.expectedProcessPuback != nil {
		return h.expectedProcessPuback(pkt)
	}
	return nil
}

func (h *mackHandler) ProcessError(err error) {
	if h.expectedError == "" {
		assert.NoError(h.t, err)
	} else {
		assert.EqualError(h.t, err, h.expectedError)
	}
}

func TestClientConnectErrorMissingAddress(t *testing.T) {
	c, err := NewClient(ClientInfo{}, &mackHandler{t: t}, nil)
	assert.EqualError(t, err, "parse : empty url")
	assert.Nil(t, c)
}

func TestClientConnectErrorWrongPort(t *testing.T) {
	cc := newConfig(t, "1234567")
	c, err := NewClient(cc, &mackHandler{t: t}, nil)
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
	c, err := NewClient(cc, &mackHandler{t: t}, nil)
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
	c, err := NewClient(cc, &mackHandler{t: t}, nil)
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
	ch := &mackHandler{t: t, expectedError: "connection refused: bad user name or password"}
	c, err := NewClient(cc, ch, nil)
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
	ch := &mackHandler{t: t, expectedError: "connection refused: not authorized"}
	c, err := NewClient(cc, ch, nil)
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
	ch := &mackHandler{t: t, expectedError: "client expected connack"}
	c, err := NewClient(cc, ch, nil)
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
	ch := &mackHandler{t: t, expectedError: "client already connecting"}
	c, err := NewClient(cc, ch, nil)
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
	c, err := NewClient(cc, &mackHandler{t: t}, nil)
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
	ch := &mackHandler{t: t, expectedError: "client missing pong"}
	c, err := NewClient(cc, ch, nil)
	assert.NoError(t, err)

	safeReceive(done)

	err = c.Close()
	assert.EqualError(t, err, "client missing pong")
}

func TestClientKeepAliveNone(t *testing.T) {
	connect := connectPacket()
	connect.KeepAlive = 0

	broker := flow.New().Debug().
		Receive(connect).
		Send(connackPacket()).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(t, port)
	cc.KeepAlive = -1
	c, err := NewClient(cc, &mackHandler{t: t}, nil)
	assert.NoError(t, err)

	<-time.After(250 * time.Millisecond)

	err = c.Close()
	assert.NoError(t, err)

	safeReceive(done)
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

	callback := func(p *packet.Publish) error {
		assert.Equal(t, "test", p.Message.Topic)
		assert.Equal(t, []byte("test"), p.Message.Payload)
		assert.Equal(t, packet.QOS(0), p.Message.QOS)
		assert.False(t, p.Message.Retain)
		close(wait)
		return nil
	}
	cc := newConfig(t, port)
	cc.Subscriptions = []TopicInfo{{Topic: "test"}}
	ch := &mackHandler{t: t, expectedProcessPublish: callback}
	c, err := NewClient(cc, ch, nil)
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

	callback1 := func(p *packet.Publish) error {
		assert.Equal(t, "test", p.Message.Topic)
		assert.Equal(t, []byte("test"), p.Message.Payload)
		assert.Equal(t, packet.QOS(1), p.Message.QOS)
		assert.False(t, p.Message.Retain)
		close(wait)
		return nil
	}
	callback2 := func(p *packet.Puback) error {
		assert.Equal(t, packet.ID(2), p.ID)
		return nil
	}
	cc := newConfig(t, port)
	cc.Subscriptions = []TopicInfo{{Topic: "test", QOS: 1}}
	ch := &mackHandler{t: t, expectedProcessPublish: callback1, expectedProcessPuback: callback2}
	c, err := NewClient(cc, ch, nil)
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
	ch := &mackHandler{t: t, expectedError: "client not connected"}
	c, err := NewClient(cc, ch, nil)
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
	ch := &mackHandler{t: t, expectedError: "client not connected"}
	c, err := NewClient(cc, ch, nil)
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
	ch := &mackHandler{t: t, expectedError: "future timeout"}
	c, err := NewClient(cc, ch, nil)
	assert.Nil(t, c)
	assert.EqualError(t, err, "future timeout")

	safeReceive(done)
}

func TestClientSubscribeFutureTimeout(t *testing.T) {
	subscribe := packet.NewSubscribe()
	subscribe.Subscriptions = []packet.Subscription{{Topic: "test"}}
	subscribe.ID = 1

	broker := flow.New().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(subscribe).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(t, port)
	cc.Timeout = time.Millisecond * 50
	cc.Subscriptions = []TopicInfo{TopicInfo{Topic: "test"}}
	ch := &mackHandler{t: t, expectedError: "failed to wait subscribe ack: future timeout"}
	c, err := NewClient(cc, ch, nil)
	assert.Nil(t, c)
	assert.EqualError(t, err, "failed to wait subscribe ack: future timeout")

	safeReceive(done)
}

func TestClientSubscribeValidate(t *testing.T) {
	subscribe := packet.NewSubscribe()
	subscribe.Subscriptions = []packet.Subscription{{Topic: "test"}}
	subscribe.ID = 1

	suback := packet.NewSuback()
	suback.ReturnCodes = []packet.QOS{packet.QOSFailure}
	suback.ID = 1

	broker := flow.New().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(subscribe).
		Send(suback).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(t, port)
	cc.ValidateSubs = true
	cc.Subscriptions = []TopicInfo{TopicInfo{Topic: "test"}}
	ch := &mackHandler{t: t, expectedError: "failed subscription"}
	c, err := NewClient(cc, ch, nil)
	assert.Nil(t, c)
	assert.EqualError(t, err, "failed to wait subscribe ack: failed subscription")

	safeReceive(done)
}

func TestClientSubscribeWithoutValidate(t *testing.T) {
	subscribe := packet.NewSubscribe()
	subscribe.Subscriptions = []packet.Subscription{{Topic: "test"}}
	subscribe.ID = 1

	suback := packet.NewSuback()
	suback.ReturnCodes = []packet.QOS{packet.QOSFailure}
	suback.ID = 1

	broker := flow.New().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(subscribe).
		Send(suback).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(t, port)
	cc.Subscriptions = []TopicInfo{TopicInfo{Topic: "test"}}
	c, err := NewClient(cc, &mackHandler{t: t}, nil)
	assert.NotNil(t, c)
	assert.NoError(t, err)

	err = c.Close()
	assert.NoError(t, err)

	safeReceive(done)
}
