// Package client implements a MQTT client and service for interacting with
// brokers.
package client

import (
	"errors"
	"fmt"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/256dpi/gomqtt/client/future"
	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/session"
	"github.com/256dpi/gomqtt/transport"

	"gopkg.in/tomb.v2"
)

// ErrClientAlreadyConnecting is returned by Connect if there has been already a
// connection attempt.
var ErrClientAlreadyConnecting = errors.New("client already connecting")

// ErrClientNotConnected is returned by Publish, Subscribe and Unsubscribe if the
// client is not currently connected.
var ErrClientNotConnected = errors.New("client not connected")

// ErrClientMissingID is returned by Connect if no ClientID has been provided in
// the config while requesting to resume a session.
var ErrClientMissingID = errors.New("client missing id")

// ErrClientConnectionDenied is returned in the Callback if the connection has
// been reject by the broker.
var ErrClientConnectionDenied = errors.New("client connection denied")

// ErrClientMissingPong is returned in the Callback if the broker did not respond
// in time to a Pingreq.
var ErrClientMissingPong = errors.New("client missing pong")

// ErrClientExpectedConnack is returned when the first received packet is not a
// Connack.
var ErrClientExpectedConnack = errors.New("client expected connack")

// ErrFailedSubscription is returned when a submitted subscription is marked as
// failed when Config.ValidateSubs must be set to true.
var ErrFailedSubscription = errors.New("failed subscription")

// A Callback is a function called by the client upon received messages or
// internal errors. An error can be returned if the callback is not already
// called with an error to instantly close the client and prevent it from
// sending any acknowledgments for the specified message.
//
// Note: Execution of the client is before the callback is called and resumed
// after the callback returns. This means that waiting on a future inside the
// callback will deadlock the client.
type Callback func(msg *packet.Message, err error) error

// A Logger is a function called by the client to log activity.
type Logger func(msg string)

const (
	clientInitialized uint32 = iota
	clientConnecting
	clientConnacked
	clientConnected
	clientDisconnecting
	clientDisconnected
)

// A Session is used to persist incoming and outgoing packets.
type Session interface {
	// NextID will return the next id for outgoing packets.
	NextID() packet.ID

	// SavePacket will store a packet in the session. An eventual existing
	// packet with the same id gets quietly overwritten.
	SavePacket(session.Direction, packet.Generic) error

	// LookupPacket will retrieve a packet from the session using a packet id.
	LookupPacket(session.Direction, packet.ID) (packet.Generic, error)

	// DeletePacket will remove a packet from the session. The method must not
	// return an error if no packet with the specified id does exists.
	DeletePacket(session.Direction, packet.ID) error

	// AllPackets will return all packets currently saved in the session.
	AllPackets(session.Direction) ([]packet.Generic, error)

	// Reset will completely reset the session.
	Reset() error
}

// A Client connects to a broker and handles the transmission of packets. It will
// automatically send PingreqPackets to keep the connection alive. Outgoing
// publish related packets will be stored in session and resent when the
// connection gets closed abruptly. All methods return Futures that get completed
// when the packets get acknowledged by the broker. Once the connection is closed
// all waiting futures get canceled.
//
// Note: If clean session is set to false and there are packets in the session,
// messages might get completed after connecting without triggering any futures
// to complete.
type Client struct {
	state uint32

	config *Config
	conn   transport.Conn

	// The session used by the client to store unacknowledged packets.
	Session Session

	// The callback to be called by the client upon receiving a message or
	// encountering an error while processing incoming packets.
	Callback Callback

	// The logger that is used to log low level information about packets
	// that have been successfully sent and received and details about the
	// automatic keep alive handler.
	Logger Logger

	clean bool

	keepAlive     time.Duration
	tracker       *Tracker
	futureStore   *future.Store
	connectFuture *future.Future

	tomb   tomb.Tomb
	mutex  sync.Mutex
	finish sync.Once
}

// New returns a new client that by default uses a fresh MemorySession.
func New() *Client {
	return &Client{
		state:       clientInitialized,
		Session:     session.NewMemorySession(),
		futureStore: future.NewStore(),
	}
}

// Connect opens the connection to the broker and sends a Connect packet. It will
// return a ConnectFuture that gets completed once a Connack has been
// received. If the Connect packet couldn't be transmitted it will return an error.
func (c *Client) Connect(config *Config) (ConnectFuture, error) {
	if config == nil {
		panic("no config specified")
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// check if already connecting
	if atomic.LoadUint32(&c.state) >= clientConnecting {
		return nil, ErrClientAlreadyConnecting
	}

	// save config
	c.config = config

	// parse url
	urlParts, err := url.ParseRequestURI(config.BrokerURL)
	if err != nil {
		return nil, err
	}

	// check client id
	if !config.CleanSession && config.ClientID == "" {
		return nil, ErrClientMissingID
	}

	// parse keep alive
	keepAlive, err := time.ParseDuration(config.KeepAlive)
	if err != nil {
		return nil, err
	}

	// allocate and initialize tracker
	c.keepAlive = keepAlive
	c.tracker = NewTracker(keepAlive)

	// dial broker (with custom dialer if present)
	if config.Dialer != nil {
		c.conn, err = config.Dialer.Dial(config.BrokerURL)
		if err != nil {
			return nil, err
		}
	} else {
		c.conn, err = transport.Dial(config.BrokerURL)
		if err != nil {
			return nil, err
		}
	}

	// set to connecting as from this point the client cannot be reused
	atomic.StoreUint32(&c.state, clientConnecting)

	// from now on the connection has been used and we have to close the
	// connection and cleanup on any subsequent error

	// save clean
	c.clean = config.CleanSession

	// reset store
	if c.clean {
		err = c.Session.Reset()
		if err != nil {
			return nil, c.cleanup(err, true, false)
		}
	}

	// allocate packet
	connect := packet.NewConnect()
	connect.ClientID = config.ClientID
	connect.KeepAlive = uint16(keepAlive.Seconds())
	connect.CleanSession = config.CleanSession

	// check for credentials
	if urlParts.User != nil {
		connect.Username = urlParts.User.Username()
		connect.Password, _ = urlParts.User.Password()
	}

	// set will
	connect.Will = config.WillMessage

	// create new ConnectFuture
	c.connectFuture = future.New()

	// send connect packet
	err = c.send(connect, false)
	if err != nil {
		return nil, c.cleanup(err, false, false)
	}

	// start process routine
	c.tomb.Go(c.processor)

	// wrap future
	wrappedFuture := &connectFuture{c.connectFuture}

	return wrappedFuture, nil
}

// Publish will send a Publish packet containing the passed parameters. It will
// return a PublishFuture that gets completed once the quality of service flow
// has been completed.
func (c *Client) Publish(topic string, payload []byte, qos packet.QOS, retain bool) (GenericFuture, error) {
	msg := &packet.Message{
		Topic:   topic,
		Payload: payload,
		QOS:     qos,
		Retain:  retain,
	}

	return c.PublishMessage(msg)
}

// PublishMessage will send a Publish containing the passed message. It will
// return a PublishFuture that gets completed once the quality of service flow
// has been completed.
func (c *Client) PublishMessage(msg *packet.Message) (GenericFuture, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// check if connected
	if atomic.LoadUint32(&c.state) != clientConnected {
		return nil, ErrClientNotConnected
	}

	// allocate publish packet
	publish := packet.NewPublish()
	publish.Message = *msg

	// set packet id
	if msg.QOS > 0 {
		publish.ID = c.Session.NextID()
	}

	// create future
	publishFuture := future.New()

	// store future
	c.futureStore.Put(publish.ID, publishFuture)

	// store packet if at least qos 1
	if msg.QOS > 0 {
		err := c.Session.SavePacket(session.Outgoing, publish)
		if err != nil {
			return nil, c.cleanup(err, true, false)
		}
	}

	// send packet
	err := c.send(publish, true)
	if err != nil {
		return nil, c.cleanup(err, false, false)
	}

	// complete and remove qos 0 future
	if msg.QOS == 0 {
		publishFuture.Complete()
		c.futureStore.Delete(publish.ID)
	}

	return publishFuture, nil
}

// Subscribe will send a Subscribe packet containing one topic to subscribe. It
// will return a SubscribeFuture that gets completed once a Suback packet has
// been received.
func (c *Client) Subscribe(topic string, qos packet.QOS) (SubscribeFuture, error) {
	return c.SubscribeMultiple([]packet.Subscription{
		{Topic: topic, QOS: qos},
	})
}

// SubscribeMultiple will send a Subscribe packet containing multiple topics to
// subscribe. It will return a SubscribeFuture that gets completed once a
// Suback packet has been received.
func (c *Client) SubscribeMultiple(subscriptions []packet.Subscription) (SubscribeFuture, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// check if connected
	if atomic.LoadUint32(&c.state) != clientConnected {
		return nil, ErrClientNotConnected
	}

	// allocate subscribe packet
	subscribe := packet.NewSubscribe()
	subscribe.ID = c.Session.NextID()
	subscribe.Subscriptions = subscriptions

	// create future
	subFuture := future.New()

	// store future
	c.futureStore.Put(subscribe.ID, subFuture)

	// send packet
	err := c.send(subscribe, true)
	if err != nil {
		return nil, c.cleanup(err, false, false)
	}

	// wrap future
	wrappedFuture := &subscribeFuture{subFuture}

	return wrappedFuture, nil
}

// Unsubscribe will send a Unsubscribe packet containing one topic to unsubscribe.
// It will return a UnsubscribeFuture that gets completed once an Unsuback packet
// has been received.
func (c *Client) Unsubscribe(topic string) (GenericFuture, error) {
	return c.UnsubscribeMultiple([]string{topic})
}

// UnsubscribeMultiple will send a Unsubscribe packet containing multiple
// topics to unsubscribe. It will return a UnsubscribeFuture that gets completed
// once an Unsuback packet has been received.
func (c *Client) UnsubscribeMultiple(topics []string) (GenericFuture, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// check if connected
	if atomic.LoadUint32(&c.state) != clientConnected {
		return nil, ErrClientNotConnected
	}

	// allocate unsubscribe packet
	unsubscribe := packet.NewUnsubscribe()
	unsubscribe.Topics = topics
	unsubscribe.ID = c.Session.NextID()

	// create future
	unsubscribeFuture := future.New()

	// store future
	c.futureStore.Put(unsubscribe.ID, unsubscribeFuture)

	// send packet
	err := c.send(unsubscribe, true)
	if err != nil {
		return nil, c.cleanup(err, false, false)
	}

	return unsubscribeFuture, nil
}

// Disconnect will send a Disconnect packet and close the connection.
//
// If a timeout is specified, the client will wait the specified amount of time
// for all queued futures to complete or cancel. If no timeout is specified it
// will not wait at all.
func (c *Client) Disconnect(timeout ...time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// check if connected
	if atomic.LoadUint32(&c.state) != clientConnected {
		return ErrClientNotConnected
	}

	// finish current packets
	if len(timeout) > 0 {
		c.futureStore.Await(timeout[0])
	}

	// set state
	atomic.StoreUint32(&c.state, clientDisconnecting)

	// send disconnect packet
	err := c.send(packet.NewDisconnect(), false)

	return c.end(err, true)
}

// Close closes the client immediately without sending a Disconnect packet and
// waiting for outgoing transmissions to finish.
func (c *Client) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// check if connected
	if atomic.LoadUint32(&c.state) < clientConnecting {
		return ErrClientNotConnected
	}

	return c.end(nil, false)
}

/* processor goroutine */

// processes incoming packets
func (c *Client) processor() error {
	first := true

	// start keep alive if greater than zero
	if c.keepAlive > 0 {
		c.tomb.Go(c.pinger)
	}

	for {
		// get next packet from connection
		pkt, err := c.conn.Receive()
		if err != nil {
			// if we are disconnecting we can ignore the error
			if atomic.LoadUint32(&c.state) >= clientDisconnecting {
				return nil
			}

			// die on any other error
			return c.die(err, false, false)
		}

		// log received message
		if c.Logger != nil {
			c.Logger(fmt.Sprintf("Received: %s", pkt.String()))
		}

		if first {
			// get connack
			connack, ok := pkt.(*packet.Connack)
			if !ok {
				return c.die(ErrClientExpectedConnack, true, false)
			}

			// process connack
			err = c.processConnack(connack)
			first = false

			// move on
			continue
		}

		// call handlers for packet types and ignore other packets
		switch typedPkt := pkt.(type) {
		case *packet.Suback:
			err = c.processSuback(typedPkt)
		case *packet.Unsuback:
			err = c.processUnsuback(typedPkt)
		case *packet.Pingresp:
			c.tracker.Pong()
		case *packet.Publish:
			err = c.processPublish(typedPkt)
		case *packet.Puback:
			err = c.processPubackAndPubcomp(typedPkt.ID)
		case *packet.Pubcomp:
			err = c.processPubackAndPubcomp(typedPkt.ID)
		case *packet.Pubrec:
			err = c.processPubrec(typedPkt.ID)
		case *packet.Pubrel:
			err = c.processPubrel(typedPkt.ID)
		}

		// return eventual error
		if err != nil {
			return err // error has already been cleaned
		}
	}
}

// handle the incoming Connack packet
func (c *Client) processConnack(connack *packet.Connack) error {
	// check state
	if atomic.LoadUint32(&c.state) != clientConnecting {
		return nil // ignore wrongly sent Connack packet
	}

	// set state
	atomic.StoreUint32(&c.state, clientConnacked)

	// fill future
	c.connectFuture.Data.Store(sessionPresentKey, connack.SessionPresent)
	c.connectFuture.Data.Store(returnCodeKey, connack.ReturnCode)

	// return connection denied error and close connection if not accepted
	if connack.ReturnCode != packet.ConnectionAccepted {
		err := c.die(ErrClientConnectionDenied, true, false)
		c.connectFuture.Cancel()
		return err
	}

	// set state to connected
	atomic.StoreUint32(&c.state, clientConnected)

	// complete future
	c.connectFuture.Complete()

	// retrieve stored packets
	packets, err := c.Session.AllPackets(session.Outgoing)
	if err != nil {
		return c.die(err, true, false)
	}

	// resend stored packets
	for _, pkt := range packets {
		// check for publish packets
		publish, ok := pkt.(*packet.Publish)
		if ok {
			// set the dup flag on a publish packet
			publish.Dup = true
		}

		// resend packet
		err = c.send(pkt, true)
		if err != nil {
			return c.die(err, false, false)
		}
	}

	return nil
}

// handle an incoming Suback packet
func (c *Client) processSuback(suback *packet.Suback) error {
	// remove packet from store
	err := c.Session.DeletePacket(session.Outgoing, suback.ID)
	if err != nil {
		return err
	}

	// get future
	subscribeFuture := c.futureStore.Get(suback.ID)
	if subscribeFuture == nil {
		return nil // ignore a wrongly sent Suback packet
	}

	// remove future from store
	c.futureStore.Delete(suback.ID)

	// validate subscriptions if requested
	if c.config.ValidateSubs {
		for _, code := range suback.ReturnCodes {
			if code == packet.QOSFailure {
				subscribeFuture.Cancel()
				return ErrFailedSubscription
			}
		}
	}

	// complete future
	subscribeFuture.Data.Store(returnCodesKey, suback.ReturnCodes)
	subscribeFuture.Complete()

	return nil
}

// handle an incoming Unsuback packet
func (c *Client) processUnsuback(unsuback *packet.Unsuback) error {
	// remove packet from store
	err := c.Session.DeletePacket(session.Outgoing, unsuback.ID)
	if err != nil {
		return err
	}

	// get future
	unsubscribeFuture := c.futureStore.Get(unsuback.ID)
	if unsubscribeFuture == nil {
		return nil // ignore a wrongly sent Unsuback packet
	}

	// complete future
	unsubscribeFuture.Complete()

	// remove future from store
	c.futureStore.Delete(unsuback.ID)

	return nil
}

// handle an incoming Publish packet
func (c *Client) processPublish(publish *packet.Publish) error {
	// call callback for unacknowledged and directly acknowledged messages
	if publish.Message.QOS <= 1 {
		if c.Callback != nil {
			err := c.Callback(&publish.Message, nil)
			if err != nil {
				return c.die(err, true, true)
			}
		}
	}

	// handle qos 1 flow
	if publish.Message.QOS == 1 {
		// prepare puback packet
		puback := packet.NewPuback()
		puback.ID = publish.ID

		// acknowledge qos 1 publish
		err := c.send(puback, true)
		if err != nil {
			return c.die(err, false, false)
		}
	}

	// handle qos 2 flow
	if publish.Message.QOS == 2 {
		// store packet
		err := c.Session.SavePacket(session.Incoming, publish)
		if err != nil {
			return c.die(err, true, false)
		}

		// prepare pubrec packet
		pubrec := packet.NewPubrec()
		pubrec.ID = publish.ID

		// acknowledge qos 2 publish
		err = c.send(pubrec, true)
		if err != nil {
			return c.die(err, false, false)
		}
	}

	return nil
}

// handle an incoming Puback or Pubcomp packet
func (c *Client) processPubackAndPubcomp(id packet.ID) error {
	// remove packet from store
	err := c.Session.DeletePacket(session.Outgoing, id)
	if err != nil {
		return err
	}

	// get future
	publishFuture := c.futureStore.Get(id)
	if publishFuture == nil {
		return nil // ignore a wrongly sent Puback or Pubcomp packet
	}

	// complete future
	publishFuture.Complete()

	// remove future from store
	c.futureStore.Delete(id)

	return nil
}

// handle an incoming Pubrec packet
func (c *Client) processPubrec(id packet.ID) error {
	// prepare pubrel packet
	pubrel := packet.NewPubrel()
	pubrel.ID = id

	// overwrite stored Publish with the Pubrel packet
	err := c.Session.SavePacket(session.Outgoing, pubrel)
	if err != nil {
		return c.die(err, true, false)
	}

	// send packet
	err = c.send(pubrel, true)
	if err != nil {
		return c.die(err, false, false)
	}

	return nil
}

// handle an incoming Pubrel packet
func (c *Client) processPubrel(id packet.ID) error {
	// get packet from store
	pkt, err := c.Session.LookupPacket(session.Incoming, id)
	if err != nil {
		return c.die(err, true, false)
	}

	// get packet from store
	publish, ok := pkt.(*packet.Publish)
	if !ok {
		return nil // ignore a wrongly sent Pubrel packet
	}

	// call callback
	if c.Callback != nil {
		err = c.Callback(&publish.Message, nil)
		if err != nil {
			return c.die(err, true, true)
		}
	}

	// prepare pubcomp packet
	pubcomp := packet.NewPubcomp()
	pubcomp.ID = publish.ID

	// acknowledge Publish packet
	err = c.send(pubcomp, true)
	if err != nil {
		return c.die(err, false, false)
	}

	// remove packet from store
	err = c.Session.DeletePacket(session.Incoming, id)
	if err != nil {
		return c.die(err, true, false)
	}

	return nil
}

/* pinger goroutine */

// manages the sending of ping packets to keep the connection alive
func (c *Client) pinger() error {
	for {
		// get current window
		window := c.tracker.Window()

		// check if ping is due
		if window < 0 {
			// check if a pong has already been sent
			if c.tracker.Pending() {
				return c.die(ErrClientMissingPong, true, false)
			}

			// send pingreq packet
			err := c.send(packet.NewPingreq(), true)
			if err != nil {
				return c.die(err, false, false)
			}

			// save ping attempt
			c.tracker.Ping()
		} else {
			// log keep alive delay
			if c.Logger != nil {
				c.Logger(fmt.Sprintf("Delay KeepAlive by %s", window.String()))
			}
		}

		select {
		case <-c.tomb.Dying():
			return tomb.ErrDying
		case <-time.After(window):
			continue
		}
	}
}

/* helpers */

// sends packet and updates lastSend
func (c *Client) send(pkt packet.Generic, async bool) error {
	// reset keep alive tracker
	c.tracker.Reset()

	// send packet
	err := c.conn.Send(pkt, async)
	if err != nil {
		return err
	}

	// log sent packet
	if c.Logger != nil {
		c.Logger(fmt.Sprintf("Sent: %s", pkt.String()))
	}

	return nil
}

// will try to cleanup as many resources as possible
func (c *Client) cleanup(err error, doClose bool, possiblyClosed bool) error {
	// cancel connect future if appropriate
	if atomic.LoadUint32(&c.state) < clientConnacked && c.connectFuture != nil {
		c.connectFuture.Cancel()
	}

	// set state
	atomic.StoreUint32(&c.state, clientDisconnected)

	// ensure that the connection gets closed
	if doClose {
		connErr := c.conn.Close()
		if connErr != nil && err == nil && !possiblyClosed {
			err = connErr
		}
	}

	// reset store
	if c.clean {
		sessErr := c.Session.Reset()
		if sessErr != nil && err == nil {
			err = sessErr
		}
	}

	// cancel all futures
	c.futureStore.Clear()

	return err
}

// used for closing and cleaning up from internal goroutines
func (c *Client) die(err error, close bool, fromCallback bool) error {
	c.finish.Do(func() {
		err = c.cleanup(err, close, false)

		if c.Callback != nil && !fromCallback {
			returnedErr := c.Callback(nil, err)
			if returnedErr == nil {
				err = nil
			}
		}
	})

	return err
}

// called by Disconnect and Close
func (c *Client) end(err error, possiblyClosed bool) error {
	// close connection
	err = c.cleanup(err, true, true)

	// shutdown goroutines
	c.tomb.Kill(nil)

	// wait for all goroutines to exit
	// goroutines will send eventual errors through the callback
	c.tomb.Wait()

	// do cleanup
	return err
}
