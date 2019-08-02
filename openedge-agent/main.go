package main

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/baidu/openedge/protocol/http"
	"github.com/baidu/openedge/protocol/mqtt"
	openedge "github.com/baidu/openedge/sdk/openedge-go"
	"github.com/baidu/openedge/utils"
)

// agent agent module
type agent struct {
	cfg  Config
	ctx  openedge.Context
	tomb utils.Tomb

	// process
	mqtt   *mqtt.Dispatcher
	events chan *Event
	// report
	certSN  string
	certKey []byte
	http    *http.Client
	// clean
	cleaner *cleaner
}

func main() {
	openedge.Run(func(ctx openedge.Context) error {
		a, err := newAgent(ctx)
		if err != nil {
			return err
		}
		defer a.close()
		err = a.start(ctx)
		if err != nil {
			return err
		}
		ctx.Wait()
		return nil
	})
}

func newAgent(ctx openedge.Context) (*agent, error) {
	var cfg Config
	err := ctx.LoadConfig(&cfg)
	if err != nil {
		return nil, err
	}
	err = defaults(&cfg)
	if err != nil {
		return nil, err
	}
	sn, err := utils.GetSerialNumber(cfg.Remote.MQTT.Cert)
	if err != nil {
		return nil, err
	}
	key, err := ioutil.ReadFile(cfg.Remote.MQTT.Key)
	if err != nil {
		return nil, err
	}
	cli, err := http.NewClient(cfg.Remote.HTTP)
	if err != nil {
		return nil, err
	}
	return &agent{
		cfg:     cfg,
		ctx:     ctx,
		http:    cli,
		events:  make(chan *Event, 1),
		certSN:  sn,
		certKey: key,
		mqtt:    mqtt.NewDispatcher(cfg.Remote.MQTT, ctx.Log()),
		cleaner: newCleaner(openedge.DefaultDBDir, path.Join(openedge.DefaultDBDir, "volumes"), ctx.Log().WithField("agent", "cleaner")),
	}, nil
}

func (a *agent) start(ctx openedge.Context) error {
	err := a.mqtt.Start(a)
	if err != nil {
		return err
	}
	return a.tomb.Go(a.reporting, a.processing)
}

func (a *agent) clean(version string) {
	a.cleaner.do(version)
}

func (a *agent) dying() <-chan struct{} {
	return a.tomb.Dying()
}

func (a *agent) close() {
	a.tomb.Kill(nil)
	a.tomb.Wait()
	a.mqtt.Close()
}

func defaults(c *Config) error {
	if c.Remote.MQTT.Address == "" {
		return fmt.Errorf("remote mqtt address missing")
	}
	if c.Remote.HTTP.CA == "" {
		return fmt.Errorf("remote http ca missing, must enable ssl")
	}
	if c.Remote.HTTP.Address == "" {
		if strings.Contains(c.Remote.MQTT.Address, "bj.baidubce.com") {
			c.Remote.HTTP.Address = "https://iotedge.bj.baidubce.com"
		} else if strings.Contains(c.Remote.MQTT.Address, "gz.baidubce.com") {
			c.Remote.HTTP.Address = "https://iotedge.gz.baidubce.com"
		} else {
			return fmt.Errorf("remote http address missing")
		}
	}
	c.Remote.Desire.Topic = fmt.Sprintf(c.Remote.Desire.Topic, c.Remote.MQTT.ClientID)
	c.Remote.Report.Topic = fmt.Sprintf(c.Remote.Report.Topic, c.Remote.MQTT.ClientID)
	c.Remote.MQTT.Subscriptions = append(c.Remote.MQTT.Subscriptions, mqtt.TopicInfo{QOS: 1, Topic: c.Remote.Desire.Topic})
	return nil
}
