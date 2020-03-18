package baetyl

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/protocol/mqtt"
	"github.com/baetyl/baetyl/utils"
)

// Context of service
type Context interface {
	context.Context

	// returns logger interface
	Log() logger.Logger
	// creates a Client that connects to the Hub through system configuration,
	// you can specify the Client ID and the topic information of the subscription.
	NewHubClient(string, []mqtt.TopicInfo) (*mqtt.Dispatcher, error)
	// LoadConfig by giving path
	LoadConfig(path string, cfg interface{}) error

	// Master RESTful API

	// inspects system stats
	//InspectSystem() (*Inspect, error)
	// gets an available port of the host
	GetAvailablePort() (string, error)

	/*
		// Master KV API

		// set kv
		SetKV(kv apiserver.KV) error
		// set kv which supports context
		SetKVConext(ctx context.Context, kv apiserver.KV) error
		// get kv
		GetKV(k []byte) (*apiserver.KV, error)
		// get kv which supports context
		GetKVConext(ctx context.Context, k []byte) (*apiserver.KV, error)
		// del kv
		DelKV(k []byte) error
		// del kv which supports context
		DelKVConext(ctx context.Context, k []byte) error
		// list kv with prefix
		ListKV(p []byte) ([]*apiserver.KV, error)
		// list kv with prefix which supports context
		ListKVContext(ctx context.Context, p []byte) ([]*apiserver.KV, error)
	*/
}

type apiConfig struct {
	Address          string `yaml:"address" json:"address"`
	TimeoutInSeconds int    `yaml:"timeout_s" json:"timeout_s"`
}

type injectConfig struct {
	Name            string          `yaml:"name" json:"name"`
	Logger          logger.LogInfo  `yaml:"logger" json:"logger"`
	CA              string          `yaml:"ca" json:"ca"`
	Certificate     string          `yaml:"certificate" json:"certificate"`
	CertificateKey  string          `yaml:"certificate_key" json:"certificate_key"`
	APIServer       apiConfig       `yaml:"apiserver" json:"apiserver"`
	LegacyAPIServer apiConfig       `yaml:"legacy_apiserver" json:"legacy_apiserver"`
	Hub             mqtt.ClientInfo `yaml:"hub" json:"hub"`
}

type ctximpl struct {
	done chan struct{}
	err  error
	cfg  injectConfig
	log  logger.Logger
	tls  *tls.Config
	m    http.Client
}

func newContext() (*ctximpl, error) {
	ctx := &ctximpl{
		done: make(chan struct{}),
	}
	err := utils.LoadYAML(configPath, &ctx.cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %s\n", err.Error())
		return nil, err
	}
	ctx.log = logger.InitLogger(ctx.cfg.Logger)

	d := filepath.Dir(configPath)
	cpath := ctx.cfg.Certificate
	if !filepath.IsAbs(cpath) {
		cpath = filepath.Join(d, cpath)
	}
	data, err := ioutil.ReadFile(cpath)
	if err != nil {
		return nil, err
	}
	var block *pem.Block
	block, data = pem.Decode(data)
	certPEM := pem.EncodeToMemory(block)
	block, data = pem.Decode(data)
	keyPEM := pem.EncodeToMemory(block)
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(data) {
		ctx.log.Errorf("append X.509 CA fail: %s", err.Error())
		return nil, errors.New("append CA fail")
	}
	ctx.tls = &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
	}
	return ctx, nil
}

func (c *ctximpl) Deadline() (deadline time.Time, ok bool) {
	ok = false
	return
}

func (c *ctximpl) Done() <-chan struct{} {
	return c.done
}

func (c *ctximpl) Err() error {
	return c.err
}

func (c *ctximpl) Value(key interface{}) interface{} {
	return nil
}

func (c *ctximpl) NewHubClient(cid string, subs []mqtt.TopicInfo) (*mqtt.Dispatcher, error) {
	if c.cfg.Hub.Address == "" {
		return nil, fmt.Errorf("hub not configured")
	}
	cc := c.cfg.Hub
	if cid != "" {
		cc.ClientID = cid
	}
	if subs != nil {
		cc.Subscriptions = subs
	}
	return mqtt.NewDispatcher(cc, c.log.WithField("cid", cid)), nil
}

func (c *ctximpl) LoadConfig(path string, cfg interface{}) error {
	return utils.LoadYAML(path, cfg)
}

func (c *ctximpl) Log() logger.Logger {
	return c.log
}

func (c *ctximpl) cancel() {
	if c.err != nil {
		return
	}
	c.err = context.Canceled
	c.done <- struct{}{}
}
