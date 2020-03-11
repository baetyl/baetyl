package mqtt

import (
	"net/url"

	"github.com/256dpi/gomqtt/transport"
	"github.com/baetyl/baetyl/utils"
)

// The Launcher helps with launching a server and accepting connections.
type Launcher struct {
	*transport.Launcher
}

// NewLauncher returns a new Launcher.
func NewLauncher(c utils.Certificate) (*Launcher, error) {
	t, err := utils.NewTLSServerConfig(c)
	if err != nil {
		return nil, err
	}
	return &Launcher{Launcher: transport.NewLauncher(transport.LaunchConfig{TLSConfig: t})}, nil
}

// Launch will launch a server based on information extracted from an URL.
func (l *Launcher) Launch(urlString string) (transport.Server, error) {
	uri, err := url.ParseRequestURI(urlString)
	if err != nil {
		return nil, err
	}
	if uri.Scheme == "ssl" {
		uri.Scheme = "tls"
	}
	return l.Launcher.Launch(uri.String())
}
