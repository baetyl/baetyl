package initz

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	gohttp "net/http"
	"os"
	"path"
	"time"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/log"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"

	"github.com/baetyl/baetyl/ami"
	"github.com/baetyl/baetyl/config"
)

type batch struct {
	name         string
	namespace    string
	securityType string
	securityKey  string
}

type Activate struct {
	log   *log.Logger
	cfg   *config.Config
	tomb  utils.Tomb
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

	ca, err := ioutil.ReadFile(cfg.Init.Active.CA)
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
	}
	for _, a := range cfg.Init.Active.Collector.Attributes {
		active.attrs[a.Name] = a.Value
	}

	active.ami, err = ami.NewAMI(cfg.AMI)
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
func (active *Activate) activate() {
	info := v1.ActiveRequest{
		BatchName:     active.batch.name,
		Namespace:     active.batch.namespace,
		SecurityType:  active.batch.securityType,
		SecurityValue: active.batch.securityKey,
		PenetrateData: active.attrs,
	}
	fv, err := active.collect()
	if err != nil {
		active.log.Error("failed to get fingerprint value", log.Error(err))
		return
	}
	if fv == "" {
		active.log.Error("fingerprint value is null", log.Error(err))
		return
	}
	info.FingerprintValue = fv
	data, err := json.Marshal(info)
	if err != nil {
		active.log.Error("failed to marshal activate info", log.Error(err))
		return
	}
	active.log.Debug("active", log.Any("info data", string(data)))

	url := fmt.Sprintf("%s%s", active.cfg.Init.Active.Address, active.cfg.Init.Active.URL)
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	resp, err := active.http.PostURL(url, bytes.NewReader(data), headers)
	if err != nil {
		active.log.Error("failed to send activate data", log.Error(err))
		return
	}
	data, err = http.HandleResponse(resp)
	if err != nil {
		active.log.Error("failed to send activate data", log.Error(err))
		return
	}
	var res v1.ActiveResponse
	err = json.Unmarshal(data, &res)
	if err != nil {
		active.log.Error("failed to unmarshal activate response data returned", log.Error(err))
		return
	}

	if err := active.genCert(res.Certificate); err != nil {
		active.log.Error("failed to create cert file", log.Error(err))
		return
	}

	active.sig <- true
}

func (active *Activate) genCert(c utils.Certificate) error {
	if err := active.createFile(active.cfg.Node.CA, []byte(c.CA)); err != nil {
		return err
	}
	if err := active.createFile(active.cfg.Node.Cert, []byte(c.Cert)); err != nil {
		return err
	}
	if err := active.createFile(active.cfg.Node.Key, []byte(c.Key)); err != nil {
		return err
	}
	return nil
}

func (active *Activate) createFile(filePath string, data []byte) error {
	dir := path.Dir(filePath)
	if !utils.DirExists(dir) {
		if err := os.Mkdir(dir, 0755); err != nil {
			return errors.Trace(err)
		}
	}
	if err := ioutil.WriteFile(filePath, data, 0755); err != nil {
		return errors.Trace(err)
	}
	return nil
}
