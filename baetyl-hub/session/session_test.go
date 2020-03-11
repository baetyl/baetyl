package session

import (
	"fmt"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/256dpi/gomqtt/packet"
	bb "github.com/baetyl/baetyl/baetyl-hub/broker"
	"github.com/baetyl/baetyl/baetyl-hub/config"
	"github.com/baetyl/baetyl/baetyl-hub/persist"
	"github.com/baetyl/baetyl/baetyl-hub/rule"
	"github.com/stretchr/testify/assert"
)

// TODO: use gomqtt's test tool: flow
func TestSessionHandle(t *testing.T) {
	r, err := prepare()
	assert.NoError(t, err)
	assert.NotNil(t, r)
	defer r.close()

	var wg sync.WaitGroup
	c := newMockClient(t, r.sessions)
	assert.NotNil(t, c)
	go func() {
		wg.Add(1)
		defer wg.Done()
		c.session.Handle()
	}()

	// connect
	c.assertOnConnectSuccess("xxx", false, nil)
	// check before subscribe
	c.assertPersistedSubscriptions(0)
	// subscribe
	c.assertOnSubscribeSuccess([]packet.Subscription{packet.Subscription{Topic: "test"}})
	// check persisted subscriptions in session.db
	c.assertPersistedSubscriptions(1)
	// publish
	c.assertPublish("test", 0, []byte{1}, false)
	// check after publish
	c.assertReceive(0, 0, "test", []byte{1}, false)
	c.close()
	wg.Wait()
}

func TestSessionConnect(t *testing.T) {
	r, err := prepare()
	assert.NoError(t, err)
	assert.NotNil(t, r)
	defer r.close()

	// Round 0
	c := newMockClient(t, r.sessions)
	assert.NotNil(t, c)
	defer c.close()
	c.assertOnConnect(t.Name(), "u1", "p1", 3, "", packet.ConnectionAccepted)

	// Round 1
	c = newMockClient(t, r.sessions)
	assert.NotNil(t, c)
	defer c.close()
	c.assertOnConnect(t.Name(), "u1", "p1", 2, "MQTT protocol version (2) invalid", packet.InvalidProtocolVersion)

	// Round 2
	c = newMockClient(t, r.sessions)
	assert.NotNil(t, c)
	defer c.close()
	c.assertOnConnect(t.Name(), "u", "p", 3, "username (u) or password not permitted", packet.BadUsernameOrPassword)
}

func TestSessionSub(t *testing.T) {
	r, err := prepare()
	assert.NoError(t, err)
	assert.NotNil(t, r)
	defer r.close()

	c := newMockClient(t, r.sessions)
	assert.NotNil(t, c)
	defer c.close()
	c.assertOnConnectSuccess(t.Name(), false, nil)
	// Round 0
	c.assertOnSubscribeSuccess([]packet.Subscription{{Topic: "test", QOS: 0}})
	// Round 1
	c.assertOnSubscribeSuccess([]packet.Subscription{{Topic: "talks"}, {Topic: "talks1", QOS: 1}, {Topic: "talks2", QOS: 1}})
	// Round 2
	c.assertOnSubscribeSuccess([]packet.Subscription{{Topic: "talks", QOS: 1}, {Topic: "talks1"}, {Topic: "talks2", QOS: 1}})
	// Round 3
	c.assertOnSubscribe([]packet.Subscription{{Topic: "test", QOS: 2}}, []packet.QOS{128})
	// Round 4
	c.assertOnSubscribe(
		[]packet.Subscription{{Topic: "talks", QOS: 2}, {Topic: "talks1", QOS: 0}, {Topic: "talks2", QOS: 1}},
		[]packet.QOS{128, 0, 1},
	)
	// Round 5
	c.assertOnSubscribe(
		[]packet.Subscription{{Topic: "talks", QOS: 2}, {Topic: "talks1", QOS: 0}, {Topic: "temp", QOS: 1}},
		[]packet.QOS{128, 0, 128},
	)
	// Round 6
	c.assertOnSubscribe(
		[]packet.Subscription{{Topic: "talks", QOS: 2}, {Topic: "talks1#/", QOS: 0}, {Topic: "talks2", QOS: 1}},
		[]packet.QOS{128, 128, 1},
	)
	// Round 7
	c.assertOnSubscribe(
		[]packet.Subscription{{Topic: "talks", QOS: 2}, {Topic: "talks1#/", QOS: 0}, {Topic: "temp", QOS: 1}},
		[]packet.QOS{128, 128, 128},
	)
}

func TestSessionPub(t *testing.T) {
	r, err := prepare()
	assert.NoError(t, err)
	assert.NotNil(t, r)
	defer r.close()

	c := newMockClient(t, r.sessions)
	assert.NotNil(t, c)
	defer c.close()
	c.assertOnConnectSuccess(t.Name(), false, nil)
	// Round 0
	c.assertOnPublish("test", 0, []byte("hello"), false)
	// Round 1
	c.assertOnPublish("test", 1, []byte("hello"), false)
	// Round 2
	c.assertOnPublishError("test", 2, []byte("hello"), "publish QOS (2) not supported")
	// Round 2
	c.assertOnPublishError("haha", 1, []byte("hello"), "publish topic (haha) not permitted")
}

func TestSessionUnSub(t *testing.T) {
	r, err := prepare()
	assert.NoError(t, err)
	assert.NotNil(t, r)
	defer r.close()

	//sub client
	subC := newMockClient(t, r.sessions)
	assert.NotNil(t, subC)
	defer subC.close()
	subC.assertOnConnectSuccess("subC", false, nil)
	//unsub
	subC.assertOnUnsubscribe([]string{"test"})
	//sub
	subC.assertOnSubscribeSuccess([]packet.Subscription{{Topic: "test"}})
	//pub client
	pubC := newMockClient(t, r.sessions)
	assert.NotNil(t, subC)
	defer subC.close()
	pubC.assertOnConnectSuccess("pubC", false, nil)
	// pub
	pubC.assertOnPublish("test", 0, []byte("hello"), false)
	// receive message
	subC.assertReceive(0, 0, "test", []byte("hello"), false)
	// unsub
	subC.assertOnUnsubscribe([]string{"test"})
	// pub
	pubC.assertOnPublish("test", 0, []byte("hello"), false)
	subC.assertReceiveTimeout()
}

func TestSessionRetain(t *testing.T) {
	r, err := prepare()
	assert.NoError(t, err)
	assert.NotNil(t, r)
	defer r.close()

	pubC := newMockClient(t, r.sessions)
	assert.NotNil(t, pubC)
	defer pubC.close()
	pubC.assertOnConnectSuccess("pubC", false, nil)
	//pub [retain = false]
	pubC.assertOnPublish("test", 1, []byte("hello"), false)
	pubC.assertRetainedMessage("", 0, "", nil)
	//pub [retain = true]
	pubC.assertOnPublish("talks", 1, []byte("hello"), true)
	pubC.assertRetainedMessage("pubC", 1, "talks", []byte("hello"))
	//sub client1
	subC1 := newMockClient(t, r.sessions)
	assert.NotNil(t, subC1)
	defer subC1.close()
	subC1.assertOnConnectSuccess("subC1", false, nil)
	// sub
	subC1.assertOnSubscribeSuccess([]packet.Subscription{{Topic: "talks"}})
	subC1.assertReceive(0, 0, "talks", []byte("hello"), true)
	subC1.assertReceiveTimeout()

	// sub client2
	subC2 := newMockClient(t, r.sessions)
	assert.NotNil(t, subC2)
	defer subC2.close()
	subC2.assertOnConnectSuccess("subC2", false, nil)
	// sub
	subC2.assertOnSubscribe([]packet.Subscription{{Topic: "talks", QOS: 1}}, []packet.QOS{1})
	subC2.assertReceive(65535, 1, "talks", []byte("hello"), true)
	subC2.assertReceiveTimeout()

	// pub remove retained message
	pubC.assertOnPublish("talks", 1, nil, true)
	pubC.assertRetainedMessage("", 0, "", nil)
	subC1.assertReceive(0, 0, "talks", nil, false) // ?
	subC2.assertReceive(3, 1, "talks", nil, false) // ?

	// sub client3
	subC3 := newMockClient(t, r.sessions)
	assert.NotNil(t, subC3)
	defer subC3.close()
	subC3.assertOnConnectSuccess("subC3", false, nil)
	// sub
	subC3.assertOnSubscribe([]packet.Subscription{{Topic: "talks", QOS: 1}}, []packet.QOS{1})
	subC3.assertReceiveTimeout()
}

func TestSessionWill_v1(t *testing.T) {
	r, err := prepare()
	assert.NoError(t, err)
	assert.NotNil(t, r)
	defer r.close()

	// Round 0 [connect without will msg]
	c := newMockClient(t, r.sessions)
	assert.NotNil(t, c)
	c.assertOnConnectSuccess(t.Name()+"_1", false, nil)
	c.assertPersistedWillMessage(nil)
	c.close()
	// Round 1 [connect with will msg, retain flag = false]
	subC := newMockClient(t, r.sessions)
	assert.NotNil(t, subC)
	defer subC.close()
	subC.assertOnConnectSuccess(t.Name()+"_sub", false, nil)
	subC.assertOnSubscribeSuccess([]packet.Subscription{{Topic: "test"}})
	// connect with will msg
	c = newMockClient(t, r.sessions)
	assert.NotNil(t, c)
	will := &packet.Message{Topic: "test", QOS: 1, Payload: []byte("hello"), Retain: false}
	c.assertOnConnectSuccess(t.Name()+"_2", false, will)
	c.assertPersistedWillMessage(will)
	// c sends will message
	c.close()
	// subC receive will message
	subC.assertReceive(0, 0, will.Topic, will.Payload, false)
}

func TestSessionWill_v2(t *testing.T) {
	r, err := prepare()
	assert.NoError(t, err)
	assert.NotNil(t, r)
	defer r.close()

	// sub will msg
	sub1C := newMockClient(t, r.sessions)
	assert.NotNil(t, sub1C)
	defer sub1C.close()
	sub1C.assertOnConnectSuccess(t.Name()+"_sub1", false, nil)
	sub1C.assertOnSubscribeSuccess([]packet.Subscription{{Topic: "test"}})

	// connect with will msg, retain flag = true
	c := newMockClient(t, r.sessions)
	assert.NotNil(t, c)
	will := &packet.Message{Topic: "test", QOS: 1, Payload: []byte("hello"), Retain: true}
	c.assertOnConnectSuccess(t.Name(), false, will)
	c.assertPersistedWillMessage(will)
	// crash
	c.close()
	c.assertRetainedMessage(t.Name(), will.QOS, will.Topic, will.Payload)
	sub1C.assertReceive(0, 0, will.Topic, will.Payload, false)

	sub2C := newMockClient(t, r.sessions)
	assert.NotNil(t, sub2C)
	defer sub2C.close()
	sub2C.assertOnConnectSuccess(t.Name()+"_sub2", false, nil)
	sub2C.assertOnSubscribeSuccess([]packet.Subscription{{Topic: "test"}})
	sub2C.assertReceive(0, 0, will.Topic, will.Payload, true)
}

func TestSessionWill_v3(t *testing.T) {
	r, err := prepare()
	assert.NoError(t, err)
	assert.NotNil(t, r)
	defer r.close()

	// sub will msg
	sub1C := newMockClient(t, r.sessions)
	assert.NotNil(t, sub1C)
	defer sub1C.close()
	sub1C.assertOnConnectSuccess(t.Name()+"_sub1", false, nil)
	sub1C.assertOnSubscribeSuccess([]packet.Subscription{{Topic: "test"}})

	// connect with will msg, retain flag = true
	c := newMockClient(t, r.sessions)
	assert.NotNil(t, c)
	defer c.close()
	will := &packet.Message{Topic: "test", QOS: 1, Payload: []byte("hello"), Retain: true}
	c.assertOnConnectSuccess(t.Name(), false, will)
	c.assertPersistedWillMessage(will)
	//session receive disconnect packet
	c.session.close(false)
	c.assertPersistedWillMessage(nil)
	c.assertRetainedMessage("", 0, "", nil)
	sub1C.assertReceiveTimeout()
}

func TestSessionWill_v4(t *testing.T) {
	r, err := prepare()
	assert.NoError(t, err)
	assert.NotNil(t, r)
	defer r.close()

	// connect with will msg, but no pub permission
	c := newMockClient(t, r.sessions)
	assert.NotNil(t, c)
	defer c.close()
	will := &packet.Message{Topic: "haha", QOS: 1, Payload: []byte("hello"), Retain: false}
	c.assertOnConnectFail(t.Name(), will)
}

func TestCleanSession(t *testing.T) {
	r, err := prepare()
	assert.NoError(t, err)
	assert.NotNil(t, r)
	defer r.close()

	pubC := newMockClient(t, r.sessions)
	assert.NotNil(t, pubC)
	defer pubC.close()
	pubC.assertOnConnect("pubC", "u2", "p2", 3, "", packet.ConnectionAccepted)

	//Round 0 [clean session = false]
	subC := newMockClient(t, r.sessions)
	assert.NotNil(t, subC)
	subC.assertOnConnectSuccess("subC", false, nil)
	subC.assertOnSubscribeSuccess([]packet.Subscription{{Topic: "test", QOS: 1}})
	assert.Len(t, subC.session.subs, 1)
	assert.Contains(t, subC.session.subs, "test")
	assert.Equal(t, packet.QOS(1), subC.session.subs["test"].QOS)

	pubC.assertOnPublish("test", 0, []byte("hello"), false)
	subC.assertReceive(0, 0, "test", []byte("hello"), false)
	subC.close()

	//Round 1 [clean session from false to false]
	subC = newMockClient(t, r.sessions)
	assert.NotNil(t, subC)
	subC.assertOnConnectSuccess("subC", false, nil)
	assert.Len(t, subC.session.subs, 1)
	assert.Contains(t, subC.session.subs, "test")
	assert.Equal(t, packet.QOS(1), subC.session.subs["test"].QOS)

	pubC.assertOnPublish("test", 1, []byte("hello"), false)
	subC.assertReceive(1, 1, "test", []byte("hello"), false)
	subC.close()

	//Round 2 [clean session from false to true]
	subC = newMockClient(t, r.sessions)
	assert.NotNil(t, subC)
	subC.assertOnConnectSuccess("subC", true, nil)
	assert.Len(t, subC.session.subs, 0)

	pubC.assertOnPublish("test", 0, []byte("hello"), false)
	subC.assertReceiveTimeout()

	subC.close()

	//Round 2 [clean session from true to false]
	subC = newMockClient(t, r.sessions)
	assert.NotNil(t, subC)
	subC.assertOnConnectSuccess("subC", false, nil)
	assert.Len(t, subC.session.subs, 0)

	pubC.assertOnPublish("test", 0, []byte("hello"), false)
	subC.assertReceiveTimeout()

	subC.close()
}

func TestConnectWithSameClientID(t *testing.T) {
	r, err := prepare()
	assert.NoError(t, err)
	assert.NotNil(t, r)
	defer r.close()

	// client1
	subC1 := newMockClient(t, r.sessions)
	assert.NotNil(t, subC1)
	subC1.assertOnConnectSuccess(t.Name(), false, nil)
	subC1.assertOnSubscribeSuccess([]packet.Subscription{{Topic: "test", QOS: 1}})

	// client 2
	subC2 := newMockClient(t, r.sessions)
	assert.NotNil(t, subC2)
	subC2.assertOnConnectSuccess(t.Name(), false, nil)
	subC2.assertOnSubscribeSuccess([]packet.Subscription{{Topic: "test", QOS: 1}})

	nop := subC1.receive()
	assert.Nil(t, nop)
	subC2.assertReceiveTimeout()

	// pub
	pubC := newMockClient(t, r.sessions)
	assert.NotNil(t, pubC)
	pubC.assertOnConnectSuccess(t.Name()+"_pub", false, nil)
	pubC.assertOnPublish("test", 1, []byte("hello"), false)

	nop = subC1.receive()
	assert.Nil(t, nop)
	subC2.assertReceive(1, 1, "test", []byte("hello"), false)
	subC2.close()
}

func TestSub0Pub0(t *testing.T) {
	testSubPub(t, 0, 0)
}

func TestSub0Pub1(t *testing.T) {
	testSubPub(t, 0, 1)
}

func TestSub1Pub0(t *testing.T) {
	testSubPub(t, 1, 0)
}

func TestSub1Pub1(t *testing.T) {
	testSubPub(t, 1, 1)
}

func testSubPub(t *testing.T, subQos packet.QOS, pubQos packet.QOS) {
	r, err := prepare()
	assert.NoError(t, err)
	assert.NotNil(t, r)
	defer r.close()

	//sub
	subC := newMockClient(t, r.sessions)
	assert.NotNil(t, subC)
	subC.assertOnConnectSuccess(t.Name()+"_sub", false, nil)
	subC.assertOnSubscribeSuccess([]packet.Subscription{{Topic: "test", QOS: subQos}})

	//pub
	pubC := newMockClient(t, r.sessions)
	assert.NotNil(t, pubC)
	pubC.assertOnConnectSuccess(t.Name()+"_pub", false, nil)
	pubC.assertOnPublish("test", pubQos, []byte("hello"), false)

	tqos := subQos
	if subQos > pubQos {
		tqos = pubQos
	}
	pkt := subC.receive()
	publish, ok := pkt.(*packet.Publish)
	assert.True(t, ok)
	assert.False(t, publish.Message.Retain)
	assert.Equal(t, "test", publish.Message.Topic)
	assert.Equal(t, tqos, publish.Message.QOS)
	assert.Equal(t, []byte("hello"), publish.Message.Payload)
	if tqos != 1 {
		assert.Equal(t, packet.ID(0), publish.ID)
		assert.Equal(t, subC.session.pids.Size(), 0)
		return
	}
	assert.Equal(t, packet.ID(1), publish.ID)
	assert.Equal(t, subC.session.pids.Size(), 1)

	// subC resends message since client does not publish ack
	pkt = subC.receive()
	publish, ok = pkt.(*packet.Publish)
	assert.True(t, ok)
	assert.True(t, publish.Dup)
	assert.False(t, publish.Message.Retain)
	assert.Equal(t, packet.ID(1), publish.ID)
	assert.Equal(t, "test", publish.Message.Topic)
	assert.Equal(t, tqos, publish.Message.QOS)
	assert.Equal(t, []byte("hello"), publish.Message.Payload)
	// subC publishes ack
	subC.assertOnPuback(1)
	subC.assertReceiveTimeout()
}

// mockClient a mock client for test
type mockClient struct {
	t       *testing.T
	i       chan packet.Generic
	o       chan packet.Generic
	session *session
}

func newMockClient(t *testing.T, m *Manager) *mockClient {
	in := make(chan packet.Generic, 10)
	out := make(chan packet.Generic, 10)
	sess := newSession(&mockCodec{in: in, out: out}, m)
	return &mockClient{
		t:       t,
		i:       in,
		o:       out,
		session: sess,
	}
}

func (c *mockClient) send(pkt packet.Generic, async bool) {
	select {
	case c.i <- pkt:
	case <-time.After(time.Second * 3):
		assert.FailNow(c.t, "send packet timeout")
	}
}

func (c *mockClient) receive() packet.Generic {
	select {
	case pkt := <-c.o:
		return pkt
	case <-time.After(time.Second * 3):
		assert.FailNow(c.t, "Receive packet timeout")
		return nil
	}
}

func (c *mockClient) close() {
	c.session.close(true)
}

func (c *mockClient) assertConnect(name, u, p string, v byte, ca packet.ConnackCode) {
	conn := &packet.Connect{ClientID: name, Username: u, Password: p, Version: v}
	c.send(conn, false)
	pkt := c.receive()
	connack := packet.Connack{ReturnCode: ca}
	assert.Equal(c.t, connack.String(), pkt.String())
}

func (c *mockClient) assertOnConnect(name, u, p string, v byte, m string, ca packet.ConnackCode) {
	conn := &packet.Connect{ClientID: name, Username: u, Password: p, Version: v}
	err := c.session.onConnect(conn)
	if m == "" {
		assert.NoError(c.t, err)
	} else {
		assert.EqualError(c.t, err, m)
	}
	pkt := c.receive()
	connack := packet.Connack{ReturnCode: ca}
	assert.Equal(c.t, connack.String(), pkt.String())
}

func (c *mockClient) assertOnConnectSuccess(name string, cs bool, will *packet.Message) {
	conn := &packet.Connect{ClientID: name, Username: "u1", Password: "p1", Version: 3, CleanSession: cs, Will: will}
	err := c.session.onConnect(conn)
	assert.NoError(c.t, err)
	pkt := c.receive()
	connack := packet.Connack{ReturnCode: packet.ConnectionAccepted}
	assert.Equal(c.t, connack.String(), pkt.String())
}

func (c *mockClient) assertOnConnectFailure(name string, will *packet.Message, e string) {
	conn := &packet.Connect{ClientID: name, Username: "u1", Password: "p1", Version: 3, Will: will}
	err := c.session.onConnect(conn)
	assert.EqualError(c.t, err, e)
}

func (c *mockClient) assertOnConnectFail(name string, will *packet.Message) {
	conn := &packet.Connect{ClientID: name, Username: "u1", Password: "p1", Version: 3, Will: will}
	err := c.session.onConnect(conn)
	assert.EqualError(c.t, err, fmt.Sprintf("will topic (%s) not permitted", will.Topic))
	pkt := c.receive()
	connack := packet.Connack{ReturnCode: packet.NotAuthorized}
	assert.Equal(c.t, connack.String(), pkt.String())
}

func (c *mockClient) assertSubscribe(subs []packet.Subscription, codes []packet.QOS) {
	sub := &packet.Subscribe{ID: 11, Subscriptions: subs}
	c.send(sub, true)
	pkt := c.receive()
	suback := packet.Suback{ReturnCodes: codes, ID: 11}
	assert.Equal(c.t, suback.String(), pkt.String())
}

func (c *mockClient) assertOnSubscribe(subs []packet.Subscription, codes []packet.QOS) {
	sub := &packet.Subscribe{ID: 12, Subscriptions: subs}
	err := c.session.onSubscribe(sub)
	assert.NoError(c.t, err)
	pkt := c.receive()
	suback := packet.Suback{ReturnCodes: codes, ID: 12}
	assert.Equal(c.t, suback.String(), pkt.String())
}

func (c *mockClient) assertOnSubscribeSuccess(subs []packet.Subscription) {
	sub := &packet.Subscribe{ID: 11, Subscriptions: subs}
	err := c.session.onSubscribe(sub)
	assert.NoError(c.t, err)
	pkt := c.receive()
	codes := make([]packet.QOS, len(subs))
	for i, s := range subs {
		codes[i] = s.QOS
	}
	suback := packet.Suback{ReturnCodes: codes, ID: 11}
	assert.Equal(c.t, suback.String(), pkt.String())
}

func (c *mockClient) assertOnUnsubscribe(topics []string) {
	unsub := &packet.Unsubscribe{Topics: topics, ID: 124}
	err := c.session.onUnsubscribe(unsub)
	assert.NoError(c.t, err)
	pkt := c.receive()
	unsuback := packet.Unsuback{ID: 124}
	assert.Equal(c.t, unsuback.String(), pkt.String())
}

func (c *mockClient) assertPublish(topic string, qos packet.QOS, payload []byte, retain bool) {
	msg := packet.Message{Topic: topic, QOS: qos, Payload: payload, Retain: retain}
	pub := &packet.Publish{Message: msg, Dup: false, ID: 123}
	c.send(pub, true)
	if qos == 1 {
		pkt := c.receive()
		puback := packet.Puback{ID: 123}
		assert.Equal(c.t, puback.String(), pkt.String())
	}
}

func (c *mockClient) assertOnPublish(topic string, qos packet.QOS, payload []byte, retain bool) {
	msg := packet.Message{Topic: topic, QOS: qos, Payload: payload, Retain: retain}
	pub := &packet.Publish{Message: msg, Dup: false, ID: 124}
	err := c.session.onPublish(pub)
	assert.NoError(c.t, err)
	if qos == 1 {
		pkt := c.receive()
		puback := packet.Puback{ID: 124}
		assert.Equal(c.t, puback.String(), pkt.String())
	}
}

func (c *mockClient) assertOnPublishError(topic string, qos packet.QOS, payload []byte, e string) {
	msg := packet.Message{Topic: topic, QOS: qos, Payload: payload}
	pub := &packet.Publish{Message: msg, Dup: false, ID: 123}
	err := c.session.onPublish(pub)
	assert.EqualError(c.t, err, e)
}

func (c *mockClient) assertOnPuback(pid packet.ID) {
	puback := packet.NewPuback()
	puback.ID = pid
	err := c.session.onPuback(puback)
	assert.NoError(c.t, err)
}

func (c *mockClient) assertPersistedSubscriptions(l int) {
	subs, err := c.session.manager.recorder.getSubs(c.session.id)
	assert.NoError(c.t, err)
	assert.Len(c.t, subs, l)
}

func (c *mockClient) assertPersistedWillMessage(expected *packet.Message) {
	will, err := c.session.manager.recorder.getWill(c.session.id)
	assert.NoError(c.t, err)
	if expected == nil {
		assert.Nil(c.t, will)
		return
	}
	assert.Equal(c.t, expected.QOS, will.QOS)
	assert.Equal(c.t, expected.Topic, will.Topic)
	assert.Equal(c.t, expected.Retain, will.Retain)
	assert.Equal(c.t, expected.Payload, will.Payload)
}

func (c *mockClient) assertRetainedMessage(cid string, qos packet.QOS, topic string, pld []byte) {
	retained, err := c.session.manager.recorder.getRetained()
	assert.NoError(c.t, err)
	if cid == "" {
		assert.Len(c.t, retained, 0)
	} else {
		assert.Len(c.t, retained, 1)
		assert.Equal(c.t, qos, retained[0].QOS)
		assert.Equal(c.t, topic, retained[0].Topic)
		assert.Equal(c.t, pld, retained[0].Payload)
		assert.True(c.t, retained[0].Retain)
	}
}

func (c *mockClient) assertReceive(pid int, qos packet.QOS, topic string, pld []byte, retain bool) {
	select {
	case pkt := <-c.o:
		assert.NotNil(c.t, pkt)
		assert.Equal(c.t, packet.PUBLISH, pkt.Type())
		p, ok := pkt.(*packet.Publish)
		assert.True(c.t, ok)
		assert.False(c.t, p.Dup)
		assert.Equal(c.t, retain, p.Message.Retain)
		assert.Equal(c.t, packet.ID(pid), p.ID)
		assert.Equal(c.t, topic, p.Message.Topic)
		assert.Equal(c.t, qos, p.Message.QOS)
		assert.Equal(c.t, pld, p.Message.Payload)
		if qos == 1 {
			c.assertOnPuback(p.ID)
		}
	case <-time.After(time.Second * 3):
		assert.FailNow(c.t, "receive packet timeout")
	}
}

func (c *mockClient) assertReceiveTimeout() {
	select {
	case pkt := <-c.o:
		assert.FailNow(c.t, "packet is not expected : %s", pkt.String())
		return
	case <-time.After(time.Second):
		return
	}
}

type mockCodec struct {
	in  chan packet.Generic
	out chan packet.Generic
}

func (c *mockCodec) Send(genericPacket packet.Generic, _ bool) error {
	select {
	case c.out <- genericPacket:
		return nil
	case <-time.After(time.Second):
		return fmt.Errorf("send packet timeout")
	}
}

func (c *mockCodec) Receive() (packet.Generic, error) {
	select {
	case pkt := <-c.in:
		return pkt, nil
	case <-time.After(time.Second):
		return nil, fmt.Errorf("send packet timeout")
	}
}

func (c *mockCodec) Conn() net.Conn {
	return nil
}

func (c *mockCodec) Close() error {
	close(c.in)
	close(c.out)
	return nil
}

func (c *mockCodec) SetReadLimit(limit int64) {}

func (c *mockCodec) SetReadTimeout(timeout time.Duration) {}

func (c *mockCodec) SetMaxWriteDelay(delay time.Duration) {}

func (c *mockCodec) SetBuffers(read, write int) {}

func (c *mockCodec) LocalAddr() net.Addr { return nil }

func (c *mockCodec) RemoteAddr() net.Addr { return nil }

func prepare() (res *resources, err error) {
	os.RemoveAll("./var")

	c, _ := config.New([]byte(""))
	c.Message.Egress.Qos1.Retry.Interval = time.Second
	c.Principals = []config.Principal{{
		Username: "u1",
		Password: "p1",
		Permissions: []config.Permission{{
			Action:  "sub",
			Permits: []string{"test", "talks", "talks1", "talks2"},
		}, {
			Action:  "pub",
			Permits: []string{"test", "talks"},
		}}}, {
		Username: "u2",
		Password: "p2",
		Permissions: []config.Permission{{
			Action:  "pub",
			Permits: []string{"test", "talks", "talks1", "talks2"},
		}}}}
	res = new(resources)
	res.factory, err = persist.NewFactory("./var/db/")
	if err != nil {
		return
	}
	res.broker, err = bb.NewBroker(c, res.factory, nil)
	if err != nil {
		return
	}
	res.rules, err = rule.NewManager(make([]config.Subscription, 0), res.broker, nil)
	if err != nil {
		return
	}
	res.sessions, err = NewManager(c, res.broker.Flow, res.rules, res.factory)
	if err != nil {
		return
	}
	res.rules.Start()
	return
}

type resources struct {
	factory  *persist.Factory
	broker   *bb.Broker
	sessions *Manager
	rules    *rule.Manager
}

func (r *resources) close() {
	if r.rules != nil {
		r.rules.Close()
	}
	if r.sessions != nil {
		r.sessions.Close()
	}
	if r.broker != nil {
		r.broker.Close()
	}
	if r.factory != nil {
		r.factory.Close()
	}
}
