package client

import (
	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/transport"
)

// Dialer defines the dialer used by a client.
type Dialer interface {
	Dial(urlString string) (transport.Conn, error)
}

// A Config holds information about establishing a connection to a broker.
type Config struct {
	// Dialer can be set to use a custom dialer.
	Dialer Dialer

	// BrokerURL is the url that is used to infer options to open the connection.
	BrokerURL string

	// ClientID can be set to the clients id.
	ClientID string

	// CleanSession can be set to request a clean session.
	CleanSession bool

	// KeepAlive should be time duration string e.g. "30s".
	KeepAlive string

	// Will message is registered on the broker upon connect if set.
	WillMessage *packet.Message

	// ValidateSubs will cause the client to fail if subscriptions failed.
	ValidateSubs bool
}

// NewConfig creates a new Config using the specified URL.
func NewConfig(url string) *Config {
	return &Config{
		BrokerURL:    url,
		CleanSession: true,
		KeepAlive:    "30s",
		ValidateSubs: true,
	}
}

// NewConfigWithClientID creates a new Config using the specified URL and client ID.
func NewConfigWithClientID(url, id string) *Config {
	config := NewConfig(url)
	config.ClientID = id
	return config
}
