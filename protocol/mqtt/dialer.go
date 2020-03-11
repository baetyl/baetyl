package mqtt

import (
	"net/url"
	"time"

	"github.com/256dpi/gomqtt/transport"
	"github.com/baetyl/baetyl/utils"
)

// The Dialer handles connecting to a server and creating a connection.
type Dialer struct {
	*transport.Dialer
}

// NewDialer returns a new Dialer.
func NewDialer(cert utils.Certificate, timeout time.Duration) (*Dialer, error) {
	tls, err := utils.NewTLSClientConfig(cert)
	if err != nil {
		return nil, err
	}
	return &Dialer{Dialer: transport.NewDialer(transport.DialConfig{TLSConfig: tls, Timeout: timeout})}, nil
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
