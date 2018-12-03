package mqtt

import (
	"crypto/tls"
	"net/url"

	"github.com/256dpi/gomqtt/transport"
	"github.com/baidu/openedge/trans"
	"github.com/juju/errors"
)

// The Launcher helps with launching a server and accepting connections.
type Launcher struct {
	TLSConfig *tls.Config
}

// NewLauncher returns a new Launcher.
func NewLauncher(c trans.Certificate) (*Launcher, error) {
	t, err := trans.NewTLSServerConfig(c.CA, c.Key, c.Cert)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &Launcher{TLSConfig: t}, nil
}

// Launch will launch a server based on information extracted from an URL.
func (l *Launcher) Launch(urlString string) (transport.Server, error) {
	urlParts, err := url.ParseRequestURI(urlString)
	if err != nil {
		return nil, err
	}

	switch urlParts.Scheme {
	case "tcp", "mqtt":
		return transport.NewNetServer(urlParts.Host)
	case "tls", "ssl", "mqtts":
		return transport.NewSecureNetServer(urlParts.Host, l.TLSConfig)
	case "ws":
		return transport.NewWebSocketServer(urlParts.Host)
	case "wss":
		return transport.NewSecureWebSocketServer(urlParts.Host, l.TLSConfig)
	}

	return nil, transport.ErrUnsupportedProtocol
}
