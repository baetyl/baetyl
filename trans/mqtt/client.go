package mqtt

import (
	"io"
	"sync"
	"time"

	"github.com/256dpi/gomqtt/client"
	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/transport"
	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/utils"
	"github.com/juju/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/tomb.v2"
)

// A Client connects to a broker and handles the transmission of packets
type Client struct {
	conn            transport.Conn
	config          ClientConfig
	tracker         *client.Tracker
	connectFuture   *Future
	subscribeFuture *Future
	callback        func(packet.Generic, error)

	finish sync.Once
	tomb   utils.Tomb
	log    *logrus.Entry
}

// NewClient returns a new client
func NewClient(cc ClientConfig, cb func(packet.Generic, error)) (*Client, error) {
	dialer, err := NewDialer(cc.Certificate)
	if err != nil {
		return nil, errors.Trace(err)
	}
	conn, err := dialer.Dial(cc.Address)
	if err != nil {
		return nil, errors.Trace(err)
	}
	c := &Client{
		conn:            conn,
		config:          cc,
		callback:        cb,
		connectFuture:   NewFuture(),
		subscribeFuture: NewFuture(),
		tracker:         client.NewTracker(cc.KeepAlive),
		log:             logger.WithFields("clientid", cc.ClientID),
	}
	err = c.connect()
	if err != nil {
		c.Close()
		return nil, errors.Trace(err)
	}
	return c, nil
}

func (c *Client) connect() (err error) {
	// allocate packet
	connect := packet.NewConnect()
	connect.ClientID = c.config.ClientID
	connect.KeepAlive = uint16(c.config.KeepAlive.Seconds())
	connect.CleanSession = c.config.CleanSession
	connect.Username = c.config.Username
	connect.Password = c.config.Password
	// connect.Will = c.config.WillMessage

	// send connect packet
	err = c.send(connect, false)
	if err != nil {
		return c.die(err)
	}

	// start process routine
	c.tomb.Go(c.processor)

	if len(c.config.Subscriptions) == 0 {
		err = c.connectFuture.Wait(c.config.Timeout)
		if err != nil {
			return c.die(err)
		}
		return nil
	}

	// allocate subscribe packet
	subscribe := packet.NewSubscribe()
	subscribe.ID = 1
	subscribe.Subscriptions = c.config.getSubscriptions()

	// send packet
	err = c.send(subscribe, true)
	if err != nil {
		return c.die(err)
	}

	err = c.connectFuture.Wait(c.config.Timeout)
	if err != nil {
		return c.die(err)
	}

	err = c.subscribeFuture.Wait(c.config.Timeout)
	if err != nil {
		return c.die(err)
	}
	return nil
}

// Send sends a generic packet
func (c *Client) Send(p packet.Generic) (err error) {
	err = c.send(p, true)
	if err != nil {
		c.die(err)
	}
	return
}

// Close closes the client immediately without sending a Disconnect packet and
// waiting for outgoing transmissions to finish.
func (c *Client) Close() error {
	c.die(nil)
	return c.tomb.Wait()
}

/* processor goroutine */

// processes incoming packets
func (c *Client) processor() error {
	c.log.Debugln("Start processor")
	defer c.log.Debugln("Stop processor")

	if c.config.KeepAlive > 0 {
		c.tomb.Go(c.pinger)
	}

	first := true

	for {
		// get next packet from connection
		pkt, err := c.conn.Receive()
		if err != nil {
			if !c.tomb.Alive() {
				return nil
			}
			if err == io.EOF {
				err = client.ErrClientNotConnected
			}
			return c.die(err)
		}

		c.log.Debugln("Received:", pkt)

		if first {
			first = false
			connack, ok := pkt.(*packet.Connack)
			if !ok {
				err = client.ErrClientExpectedConnack
				return c.die(err)
			}

			if connack.ReturnCode != packet.ConnectionAccepted {
				err = errors.New(connack.ReturnCode.String())
				return c.die(err)
			}

			c.connectFuture.Complete()
			continue
		}

		switch p := pkt.(type) {
		case *packet.Pingresp:
			c.tracker.Pong()
		case *packet.Connack:
			err = client.ErrClientAlreadyConnecting
			return c.die(err)
		case *packet.Suback:
			c.subscribeFuture.Complete()
			for _, code := range p.ReturnCodes {
				if code == packet.QOSFailure {
					err = client.ErrFailedSubscription
					return c.die(err)
				}
			}
		default:
			if c.callback != nil {
				c.callback(pkt, nil)
			}
		}
	}
}

/* pinger goroutine */

// manages the sending of ping packets to keep the connection alive
func (c *Client) pinger() (err error) {
	c.log.Debugln("Start pinger")
	defer c.log.Debugln("Stop pinger")

	for {
		// get current window
		window := c.tracker.Window()

		// check if ping is due
		if window < 0 {
			// check if a pong has already been sent
			if c.tracker.Pending() {
				err = client.ErrClientMissingPong
				return c.die(err)
			}

			// send pingreq packet
			err = c.send(packet.NewPingreq(), false)
			if err != nil {
				return c.die(err)
			}

			// save ping attempt
			c.tracker.Ping()
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
		return errors.Trace(err)
	}

	// config.Logger sent packet
	c.log.Debugln("Sent:", pkt)

	return nil
}

// used for closing and cleaning up from internal goroutines
func (c *Client) die(err error) error {
	c.finish.Do(func() {
		if err == nil {
			c.send(packet.NewDisconnect(), false)
		} else {
			c.callback(nil, err)
			c.log.WithError(err).Errorln("MQTT client raises error")
		}
		c.tomb.Kill(err)
		c.connectFuture.Cancel(err)
		c.subscribeFuture.Cancel(err)
		if c.conn != nil {
			c.conn.Close()
		}
	})
	return errors.Trace(err)
}
