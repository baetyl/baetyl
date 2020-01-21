package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/baetyl/baetyl/baetyl-agent/common"
	"github.com/baetyl/baetyl/baetyl-agent/config"
	"github.com/baetyl/baetyl/protocol/http"
	"github.com/baetyl/baetyl/protocol/mqtt"
	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
	"github.com/baetyl/baetyl/utils"
)

// agent agent module
type agent struct {
	cfg  config.Config
	ctx  baetyl.Context
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
	// active
	node  *node
	batch *batch
	srv   *Server
	attrs map[string]string
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
		err = a.start()
		if err != nil {
			return err
		}
		ctx.Wait()
		return nil
	})
}

func newAgent(ctx baetyl.Context) (*agent, error) {
	var cfg config.Config
	err := ctx.LoadConfig(&cfg)
	if err != nil {
		return nil, err
	}
	sn := ""
	key := []byte{}
	var dispatcher *mqtt.Dispatcher
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
		dispatcher = mqtt.NewDispatcher(*cfg.Remote.MQTT, ctx.Log())
	}
	a := &agent{
		cfg:     cfg,
		ctx:     ctx,
		events:  make(chan *Event, 1),
		certSN:  sn,
		certKey: key,
		mqtt:    dispatcher,
		cleaner: newCleaner(baetyl.DefaultDBDir, path.Join(baetyl.DefaultDBDir, "volumes"), ctx.Log().WithField("agent", "cleaner")),
	}
	a.http, err = http.NewClient(*cfg.Remote.HTTP)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (a *agent) start() error {
	nodeName := os.Getenv(common.NodeName)
	nodeNamespace := os.Getenv(common.NodeNamespace)
	if nodeName != "" && nodeNamespace != "" {
		a.node = &node{
			Name:      nodeName,
			Namespace: nodeNamespace,
		}
	} else {
		batchName := os.Getenv(common.BatchName)
		batchNamespace := os.Getenv(common.BatchNamespace)
		a.batch = &batch{
			Name:      batchName,
			Namespace: batchNamespace,
		}
	}
	if a.mqtt != nil {
		err := a.mqtt.Start(a)
		if err != nil {
			return err
		}
	}
	if a.cfg.Server.Listen != "" && a.batch != nil {
		err := a.NewServer(a.cfg.Server, a.ctx.Log())
		if err != nil {
			return err
		}
		err = a.StartServer()
		if err != nil {
			return err
		}

	} else {
		err := a.tomb.Go(a.reporting, a.processing)
		if err != nil {
			return err
		}
	}
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
	if a.srv != nil {
		a.CloseServer()
	}
}

func defaults(c *config.Config) error {
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
