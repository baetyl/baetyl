package initz

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"math/rand"
	gohttp "net/http"
	"os"
	"time"

	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/log"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	v2utils "github.com/baetyl/baetyl-go/v2/utils"

	"github.com/baetyl/baetyl/v2/ami"
	"github.com/baetyl/baetyl/v2/config"
	"github.com/baetyl/baetyl/v2/utils"
)

const (
	LinkMqtt = "mqttlink"
)

type batch struct {
	name         string
	namespace    string
	securityType string
	securityKey  string
	mode         string
}

type Activate struct {
	log   *log.Logger
	cfg   *config.Config
	tomb  v2utils.Tomb
	http  *http.Client
	srv   *gohttp.Server
	ami   ami.AMI
	batch *batch
	attrs map[string]string
	sig   chan bool
}

// NewActivate creates a new activate
func NewActivate(cfg *config.Config) (*Activate, error) {
	// TODO 优化ToClientOptions 支持只配置ca的单向认证
	ops, err := cfg.Init.Active.ToClientOptions()
	if err != nil {
		return nil, errors.Trace(err)
	}

	ca, err := os.ReadFile(cfg.Init.Active.CA)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(ca)
	ops.TLSConfig = &tls.Config{RootCAs: pool}
	ops.TLSConfig.InsecureSkipVerify = cfg.Init.Active.InsecureSkipVerify

	active := &Activate{
		cfg:   cfg,
		sig:   make(chan bool, 1),
		http:  http.NewClient(ops),
		attrs: map[string]string{},
		log:   log.With(log.Any("init", "active")),
	}
	active.batch = &batch{
		name:         cfg.Init.Batch.Name,
		namespace:    cfg.Init.Batch.Namespace,
		securityType: cfg.Init.Batch.SecurityType,
		securityKey:  cfg.Init.Batch.SecurityKey,
		mode:         cfg.Init.Batch.Mode,
	}
	for _, a := range cfg.Init.Active.Collector.Attributes {
		active.attrs[a.Name] = a.Value
	}

	active.ami, err = ami.NewAMI(context.RunMode(), cfg.AMI)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return active, nil
}

func (active *Activate) Start() {
	if active.cfg.Init.Active.Collector.Server.Listen == "" {
		err := active.tomb.Go(active.activating)
		if err != nil {
			active.log.Error("failed to start report and process routine", log.Error(err))
			return
		}
	} else {
		err := active.tomb.Go(active.startServer)
		if err != nil {
			active.log.Error("active", log.Error(err))
		}
	}
}

func (active *Activate) Close() {
	if active.srv != nil {
		active.closeServer()
	}
	active.tomb.Kill(nil)
	active.tomb.Wait()
}

func (active *Activate) WaitAndClose() {
	if _, ok := <-active.sig; !ok {
		active.log.Error("Activate get sig error")
	}
	active.Close()
}

func (active *Activate) activating() error {
	active.activate()
	t := time.NewTicker(active.cfg.Init.Active.Interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
			active.activate()
		case <-active.tomb.Dying():
			return nil
		}
	}
}

// Report reports info
func (active *Activate) activate() error {
	info := v1.ActiveRequest{
		BatchName:     active.batch.name,
		Namespace:     active.batch.namespace,
		SecurityType:  active.batch.securityType,
		SecurityValue: active.batch.securityKey,
		Mode:          active.batch.mode,
		PenetrateData: active.attrs,
	}
	fv, err := active.collect()
	if err != nil {
		active.log.Error("failed to get fingerprint value", log.Error(err))
		return err
	}
	if fv == "" {
		active.log.Error("fingerprint value is null", log.Error(err))
		return errors.New("fingerprint value is null")
	}
	info.FingerprintValue = fv
	data, err := json.Marshal(info)
	if err != nil {
		active.log.Error("failed to marshal activate info", log.Error(err))
		return err
	}
	active.log.Debug("active", log.Any("info data", string(data)))

	url := fmt.Sprintf("%s%s", active.cfg.Init.Active.Address, active.cfg.Init.Active.URL)
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	resp, err := active.http.PostURL(url, bytes.NewReader(data), headers)
	if err != nil {
		active.log.Error("failed to send activate data", log.Error(err))
		return err
	}

	data, err = http.HandleResponse(resp)
	if err != nil {
		active.log.Error("failed to send activate data", log.Error(err))
		return err
	}
	var res v1.ActiveResponse
	err = json.Unmarshal(data, &res)
	if err != nil {
		active.log.Error("failed to unmarshal activate response data returned", log.Error(err))
		return err
	}

	if err = genCert(&active.cfg.Node, res.Certificate); err != nil {
		active.log.Error("failed to create cert file", log.Error(err))
		return err
	}

	if active.cfg.Plugin.Link == LinkMqtt {
		if err = genCert(&active.cfg.MqttLink.Cert, res.MqttCert); err != nil {
			active.log.Error("failed to create mqtt cert file", log.Error(err))
			return err
		}
	}

	active.sig <- true
	return nil
}

func genCert(path *v2utils.Certificate, c v2utils.Certificate) error {
	if err := utils.CreateWriteFile(path.CA, []byte(c.CA)); err != nil {
		return err
	}
	if err := utils.CreateWriteFile(path.Cert, []byte(c.Cert)); err != nil {
		return err
	}
	if err := utils.CreateWriteFile(path.Key, []byte(c.Key)); err != nil {
		return err
	}
	return nil
}
