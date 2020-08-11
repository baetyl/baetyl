package httplink

import (
	"crypto/x509"
	"encoding/json"
	"os"
	"strings"

	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/log"
	v2plugin "github.com/baetyl/baetyl-go/v2/plugin"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/baetyl/baetyl/plugin"
)

const (
	EnvKeyNodeNamespace = "BAETYL_NODE_NAMESPACE"
	DefaultConfFile     = "etc/baetyl/service.yml"
)

func init() {
	v2plugin.RegisterFactory("httplink", New)
}

type httpLink struct {
	cfg  Config
	http *http.Client
	log  *log.Logger
}

func (l *httpLink) Close() error {
	return nil
}

func New() (v2plugin.Plugin, error) {
	var cfg Config
	if err := utils.LoadYAML(DefaultConfFile, &cfg); err != nil {
		return nil, errors.Trace(err)
	}
	ops, err := cfg.HTTPLink.HTTP.ToClientOptions()
	if err != nil {
		return nil, errors.Trace(err)
	}
	if ops.TLSConfig == nil {
		return nil, errors.Trace(plugin.ErrLinkTLSConfigMissing)
	}
	link := &httpLink{
		cfg:  cfg,
		http: http.NewClient(ops),
		log:  log.With(log.Any("plugin", "httplink")),
	}
	if len(ops.TLSConfig.Certificates) == 1 && len(ops.TLSConfig.Certificates[0].Certificate) == 1 {
		cert, err := x509.ParseCertificate(ops.TLSConfig.Certificates[0].Certificate[0])
		if err == nil {
			res := strings.SplitN(cert.Subject.CommonName, ".", 2)
			if len(res) != 2 || res[0] == "" || res[1] == "" {
				link.log.Error("failed to parse node name from cert")
			} else {
				os.Setenv(context.EnvKeyNodeName, res[1])
				os.Setenv(EnvKeyNodeNamespace, res[0])
			}
		} else {
			link.log.Error("certificate format error")
		}
	}
	return link, nil
}

func (l *httpLink) Receive() (<-chan *specv1.Message, <-chan error) {
	return nil, nil
}

func (l *httpLink) Request(msg *specv1.Message) (*specv1.Message, error) {
	l.log.Debug("http link send request message", log.Any("message", msg))
	pld, err := json.Marshal(msg.Content)
	if err != nil {
		return nil, errors.Trace(err)
	}
	var data []byte
	res := &specv1.Message{Kind: msg.Kind}
	switch msg.Kind {
	case specv1.MessageReport:
		data, err = l.http.PostJSON(l.cfg.HTTPLink.ReportURL, pld)
		if err != nil {
			return nil, errors.Trace(err)
		}
		var desire specv1.Desire
		if err = json.Unmarshal(data, &desire); err != nil {
			return nil, errors.Trace(err)
		}
		res.Content = desire
	case specv1.MessageDesire:
		data, err = l.http.PostJSON(l.cfg.HTTPLink.DesireURL, pld)
		if err != nil {
			return nil, errors.Trace(err)
		}
		var desireRes specv1.DesireResponse
		if err = json.Unmarshal(data, &desireRes); err != nil {
			return nil, errors.Trace(err)
		}
		res.Content = desireRes
	default:
		return nil, errors.Errorf("unsupported message kind")
	}
	return res, nil
}

func (l *httpLink) Send(msg *specv1.Message) error {
	return nil
}

func (l *httpLink) IsAsyncSupported() bool {
	return false
}
