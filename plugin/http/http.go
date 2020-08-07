package http

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl/plugin"
	"os"
	"strings"
)

var (
	// ErrSyncTLSConfigMissing certificate bidirectional authentication is required for connection with cloud
	ErrSyncTLSConfigMissing = fmt.Errorf("certificate bidirectional authentication is required for connection with cloud")
)

const (
	EnvKeyNodeName      = "BAETYL_NODE_NAME"
	EnvKeyNodeNamespace = "BAETYL_NODE_NAMESPACE"
)

type httpLink struct {
	http *http.Client
	log  *log.Logger
}

func (l *httpLink) Close() error {
	return nil
}

func init() {
	plugin.RegisterFactory("link", New)
}

func New() (plugin.Plugin, error) {
	var cfg Config
	ops, err := cfg.HTTP.ToClientOptions()
	if err != nil {
		return nil, errors.Trace(err)
	}
	if ops.TLSConfig == nil {
		return nil, errors.Trace(ErrSyncTLSConfigMissing)
	}
	link := &httpLink{
		http: http.NewClient(ops),
		log:  log.With(log.Any("plugin", "http")),
	}
	if len(ops.TLSConfig.Certificates) == 1 && len(ops.TLSConfig.Certificates[0].Certificate) == 1 {
		cert, err := x509.ParseCertificate(ops.TLSConfig.Certificates[0].Certificate[0])
		if err == nil {
			res := strings.SplitN(cert.Subject.CommonName, ".", 2)
			if len(res) != 2 || res[0] == "" || res[1] == "" {
				link.log.Error("failed to parse node name from cert")
			} else {
				os.Setenv(EnvKeyNodeName, res[1])
				os.Setenv(EnvKeyNodeNamespace, res[0])
			}
		} else {
			link.log.Error("certificate format error")
		}
	}
	return link, nil
}

func (l *httpLink) Receive(msg *plugin.Message) error {
	return nil
}
func (l *httpLink) Request(msg *plugin.Message) (*plugin.Message, error) {
	pld, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Trace(err)
	}
	l.log.Debug("http link send request message", log.Any("message", string(pld)))
	data, err := l.http.PostJSON(msg.URI, pld)
	if err != nil {
		return nil, errors.Trace(err)
	}
	var res plugin.Message
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &res, nil
}

func (l *httpLink) Send(msg *plugin.Message) error {
	return nil
}

func (l *httpLink) IsAsyncSupported() bool {
	return false
}
