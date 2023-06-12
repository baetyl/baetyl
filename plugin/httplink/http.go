package httplink

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/log"
	goplugin "github.com/baetyl/baetyl-go/v2/plugin"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"

	"github.com/baetyl/baetyl/v2/initz"
	"github.com/baetyl/baetyl/v2/plugin"
)

func init() {
	goplugin.RegisterFactory("httplink", New)
}

type httpLink struct {
	cfg   Config
	addrs []string
	ops   *http.ClientOptions
	http  *http.Client
	log   *log.Logger
}

func (l *httpLink) Close() error {
	return nil
}

func New() (goplugin.Plugin, error) {
	var cfg Config
	if err := utils.LoadYAML(plugin.ConfFile, &cfg); err != nil {
		return nil, errors.Trace(err)
	}
	// to use node cert
	cfg.HTTPLink.HTTP.Certificate.CA = cfg.Node.CA
	cfg.HTTPLink.HTTP.Certificate.Key = cfg.Node.Key
	cfg.HTTPLink.HTTP.Certificate.Cert = cfg.Node.Cert
	ops, err := cfg.HTTPLink.HTTP.ToClientOptions()
	if err != nil {
		return nil, errors.Trace(err)
	}
	if ops.TLSConfig == nil {
		return nil, errors.Trace(plugin.ErrLinkTLSConfigMissing)
	}
	link := &httpLink{
		cfg:   cfg,
		ops:   ops,
		addrs: strings.Split(cfg.HTTPLink.HTTP.Address, ","),
		http:  http.NewClient(ops),
		log:   log.With(log.Any("plugin", "httplink")),
	}
	if addrEnv := os.Getenv(initz.KeyBaetylSyncAddr); addrEnv != "" {
		link.addrs = strings.Split(addrEnv, ",")
	}
	return link, nil
}

func (l *httpLink) Receive() (<-chan *specv1.Message, <-chan error) {
	return nil, nil
}

// Request 现在 http link 支持以下类型消息的处理
// 上报类：上报 report(默认)、设备消息上报 deviceReport、设备生命周期上报 thing.lifecycle.post
// 同步类：期望 desire(默认)、设备消息同步 deviceDesire
func (l *httpLink) Request(msg *specv1.Message) (*specv1.Message, error) {
	l.log.Debug("http link send request", log.Any("message", msg))
	pld, err := json.Marshal(msg.Content)
	if err != nil {
		return nil, errors.Trace(err)
	}
	var data []byte
	res := &specv1.Message{Kind: msg.Kind}
	if msg.Metadata == nil {
		msg.Metadata = map[string]string{}
	}
	msg.Metadata["kind"] = string(msg.Kind)
	switch msg.Kind {
	case specv1.MessageReport, specv1.MessageDeviceReport, specv1.MessageDeviceLifecycleReport:
		data, err = l.post(l.cfg.HTTPLink.ReportURL, pld, msg.Metadata)
		if err != nil {
			return nil, errors.Trace(err)
		}
	case specv1.MessageDesire, specv1.MessageDeviceDesire:
		data, err = l.post(l.cfg.HTTPLink.DesireURL, pld, msg.Metadata)
		if err != nil {
			return nil, errors.Trace(err)
		}
	default:
		return nil, errors.Errorf("unsupported message kind")
	}
	data, err = utils.ParseEnv(data)
	if err != nil {
		return nil, errors.Trace(err)
	}
	res.Content.SetJSON(data)
	l.log.Debug("http link receive response", log.Any("message", res))
	return res, nil
}

func (l *httpLink) State() *specv1.Message {
	return nil
}

func (l *httpLink) Send(msg *specv1.Message) error {
	return nil
}

func (l *httpLink) IsAsyncSupported() bool {
	return false
}

func (l *httpLink) post(url string, pld []byte, headers map[string]string) ([]byte, error) {
	errs := []string{}
	for _, addr := range l.addrs {
		l.ops.Address = addr
		data, err := l.http.PostJSON(url, pld, headers)
		if err != nil {
			l.log.Warn("post error", log.Any("addr", addr), log.Error(err))
			errs = append(errs, err.Error())
		} else {
			return data, nil
		}
	}
	return nil, errors.Trace(errors.New(strings.Join(errs, ";")))
}
