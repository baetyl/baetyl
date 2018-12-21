package client

import (
	"time"

	"github.com/256dpi/gomqtt/packet"
)

// ClearSession will connect to the specified broker and request a clean session.
func ClearSession(config *Config, timeout time.Duration) error {
	// copy config
	newConfig := *config
	newConfig.CleanSession = true

	// create client
	client := New()

	// connect to broker
	future, err := client.Connect(&newConfig)
	if err != nil {
		return err
	}

	// wait for future
	err = future.Wait(timeout)
	if err != nil {
		return err
	}

	// disconnect
	err = client.Disconnect()
	if err != nil {
		return err
	}

	return nil
}

// PublishMessage will connect to the specified broker to publish the passed message.
func PublishMessage(config *Config, msg *packet.Message, timeout time.Duration) error {
	// create client
	client := New()

	// connect to broker
	future, err := client.Connect(config)
	if err != nil {
		return err
	}

	// wait on future
	err = future.Wait(timeout)
	if err != nil {
		return err
	}

	// publish message
	publishFuture, err := client.PublishMessage(msg)
	if err != nil {
		return err
	}

	// wait on future
	err = publishFuture.Wait(timeout)
	if err != nil {
		return err
	}

	// disconnect
	err = client.Disconnect()
	if err != nil {
		return err
	}

	return nil
}

// ClearRetainedMessage will connect to the specified broker and send an empty
// retained message to force any already retained message to be cleared.
func ClearRetainedMessage(config *Config, topic string, timeout time.Duration) error {
	return PublishMessage(config, &packet.Message{
		Topic:   topic,
		Payload: nil,
		QOS:     0,
		Retain:  true,
	}, timeout)
}

// ReceiveMessage will connect to the specified broker and issue a subscription
// for the specified topic and return the first message received.
func ReceiveMessage(config *Config, topic string, qos packet.QOS, timeout time.Duration) (*packet.Message, error) {
	// create client
	client := New()

	// connect to broker
	future, err := client.Connect(config)
	if err != nil {
		return nil, err
	}

	// wait for future
	err = future.Wait(timeout)
	if err != nil {
		return nil, err
	}

	// create channel
	msgCh := make(chan *packet.Message)
	errCh := make(chan error)

	// set callback
	client.Callback = func(msg *packet.Message, err error) error {
		if err != nil {
			errCh <- err
			return nil
		}

		msgCh <- msg
		return nil
	}

	// make subscription
	subscribeFuture, err := client.Subscribe(topic, qos)
	if err != nil {
		return nil, err
	}

	// wait for future
	err = subscribeFuture.Wait(timeout)
	if err != nil {
		return nil, err
	}

	// prepare message
	var msg *packet.Message

	// wait for error, message or timeout
	select {
	case err = <-errCh:
		return nil, err
	case msg = <-msgCh:
	case <-time.After(timeout):
	}

	// disconnect
	err = client.Disconnect()
	if err != nil {
		return nil, err
	}

	return msg, nil
}
