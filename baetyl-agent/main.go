package main

import (
	"fmt"
	"github.com/baetyl/baetyl-go/link"
	"github.com/baetyl/baetyl/baetyl-agent/common"
	"github.com/baetyl/baetyl/protocol/http"
	"github.com/baetyl/baetyl/protocol/mqtt"
	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
	"github.com/baetyl/baetyl/utils"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

// agent agent module
type agent struct {
	cfg  Config
	ctx  baetyl.Context
	tomb utils.Tomb

	// process
	mqtt   *mqtt.Dispatcher
	events chan *Event
	infos  chan *BackwardInfo
	// report
	certSN  string
	certKey []byte
	http    *http.Client
	link    *link.Client
	// clean
	cleaner         *cleaner
	shadowName      string
	shadowNamespace string
}

func main() {
	// backward compatibility
	addSymlinkCompatible()
	baetyl.Run(func(ctx baetyl.Context) error {
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

func newAgent(ctx baetyl.Context) (*agent, error) {
	var cfg Config
	err := ctx.LoadConfig(&cfg)
	if err != nil {
		return nil, err
	}
	sn := ""
	key := []byte{}
	if cfg.Remote.MQTT != nil {
		err = defaults(&cfg)
		if err != nil {
			return nil, err
		}
		sn, err = utils.GetSerialNumber(cfg.Remote.MQTT.Cert)
		if err != nil {
			return nil, err
		}
		key, err = ioutil.ReadFile(cfg.Remote.MQTT.Key)
		if err != nil {
			return nil, err
		}
	}
	a := &agent{
		cfg:     cfg,
		ctx:     ctx,
		events:  make(chan *Event, 1),
		infos:   make(chan *BackwardInfo, 1),
		certSN:  sn,
		certKey: key,
		cleaner: newCleaner(baetyl.DefaultDBDir, path.Join(baetyl.DefaultDBDir, "volumes"), ctx.Log().WithField("agent", "cleaner")),
	}

	a.shadowName = os.Getenv(common.ShadowName)
	a.shadowNamespace = os.Getenv(common.ShadowNamespace)
	if cfg.Remote.Link != nil && (a.shadowName == "" || a.shadowNamespace == "") {
		return nil, fmt.Errorf("can not report info without shadow name or namespace: %s", err.Error())
	}

	if cfg.Remote.HTTP != nil {
		a.http, err = http.NewClient(*cfg.Remote.HTTP)
	} else {
		a.http, err = http.NewClient(http.ClientInfo{Timeout: 5 * time.Minute, KeepAlive: 10 * time.Minute})
	}
	if err != nil {
		return nil, err
	}
	if cfg.Remote.MQTT != nil {
		a.mqtt = mqtt.NewDispatcher(*cfg.Remote.MQTT, ctx.Log())
	}
	if cfg.Remote.Link != nil {
		a.link, err = link.NewClient(*cfg.Remote.Link, nil)
		if err != nil {
			return nil, err
		}
	}
	return a, nil
}

func (a *agent) start(ctx baetyl.Context) error {
	if a.mqtt != nil {
		err := a.mqtt.Start(a)
		if err != nil {
			return err
		}
		err = a.tomb.Go(a.processing)
		if err != nil {
			return err
		}
	}
	if a.link != nil {
		err := a.tomb.Go(a.linkProcess)
		if err != nil {
			return err
		}
	}
	err := a.tomb.Go(a.reporting)
	if err != nil {
		return err
	}
	//if a.link != nil {
	//	err := a.tomb.Go(a.linkReporting)
	//	if err != nil {
	//		return err
	//	}
	//}
	return nil
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
	if a.mqtt != nil {
		a.mqtt.Close()
	}
	if a.link != nil {
		a.link.Close()
	}
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

func addSymlinkCompatible() {
	list := map[string]string{
		// openedge -> baetyl
		baetyl.DefaultMasterConfDir: baetyl.PreviousMasterConfDir,
		baetyl.DefaultDBDir:         baetyl.PreviousDBDir,
		baetyl.DefaultLogDir:        baetyl.PreviousLogDir,
		// baetyl -> openedge
		baetyl.PreviousMasterConfDir: baetyl.DefaultMasterConfDir,
		baetyl.PreviousDBDir:         baetyl.DefaultDBDir,
		baetyl.PreviousLogDir:        baetyl.DefaultLogDir,
	}
	for k, v := range list {
		err := addSymlink(k, v)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}

func addSymlink(src, desc string) error {
	if utils.PathExists(src) {
		err := utils.CreateSymlink(path.Base(src), desc)
		if err != nil {
			return err
		}
	}
	return nil
}
