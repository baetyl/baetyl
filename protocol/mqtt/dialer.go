package mqtt

import (
	"crypto/tls"
	"net/url"

	"github.com/256dpi/gomqtt/transport"
	"github.com/baetyl/baetyl/utils"
)

// The Dialer handles connecting to a server and creating a connection.
type Dialer struct {
	*transport.Dialer
}

// NewDialer returns a new Dialer.
func NewDialer(c utils.Certificate) (*Dialer, error) {
	return NewDialer2(nil, c)
}

// NewDialer2 returns a new Dialer.
func NewDialer2(c *tls.Config, c2 utils.Certificate) (*Dialer, error) {
	if c == nil {
		var err error
		c, err = utils.NewTLSClientConfig(c2)
		if err != nil {
			return nil, err
		}
	}
	d := &Dialer{Dialer: transport.NewDialer()}
	d.TLSConfig = c
	return d, nil
}

// Dial initiates a connection based in information extracted from an URL.
func (d *Dialer) Dial(urlString string) (transport.Conn, error) {
	uri, err := url.ParseRequestURI(urlString)
	if err != nil {
		return nil, err
	}
	if uri.Scheme == "ssl" {
		uri.Scheme = "tls"
	}
	return d.Dialer.Dial(uri.String())
}
