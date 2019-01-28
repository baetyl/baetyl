package sdk

import (
	"errors"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/256dpi/gomqtt/packet"
	openedge "github.com/baidu/openedge/api/go"
	"github.com/baidu/openedge/protocol/jrpc"
	"github.com/baidu/openedge/protocol/mqtt"
	"github.com/baidu/openedge/utils"
)

// DefaultConfigPath is the path to config of this service
const DefaultConfigPath = "etc/openedge/service.yml"

type context struct {
	cfg    openedge.Config
	topic  string
	handle func(*openedge.Message) error
	hub    *mqtt.Dispatcher
	master *jrpc.Client
	log    openedge.Logger
}

func (c *context) Config() *openedge.Config {
	return &c.cfg
}

func (c *context) Log() openedge.Logger {
	return c.log
}

func (c *context) WaitExit() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	signal.Ignore(syscall.SIGPIPE)
	<-sig
}

func (c *context) Subscribe(topic openedge.TopicInfo, handle func(*openedge.Message) error) error {
	if c.hub == nil {
		return errors.New("no hub")
	}
	p := packet.NewSubscribe()
	p.Subscriptions = []packet.Subscription{
		packet.Subscription{
			Topic: topic.Topic,
			QOS:   packet.QOS(topic.QoS),
		},
	}
	p.ID = 1
	err := c.hub.Send(p)
	if err != nil {
		return err
	}
	// FIXME not support multiple subscription
	c.topic = topic.Topic
	c.handle = handle
	return nil
}

func (c *context) SendMessage(message *openedge.Message) error {
	if c.hub == nil {
		return errors.New("no hub")
	}
	p := packet.NewPublish()
	p.Message.Topic = message.Topic
	p.Message.QOS = packet.QOS(message.QoS)
	p.Message.Payload = message.Payload
	return c.hub.Send(p)
}

func (c *context) StartService(info *openedge.ServiceInfo, config []byte) error {
	var reply openedge.StartServiceResponse
	err := c.master.Call(
		openedge.CallStartService,
		openedge.StartServiceRequest{
			Info:   *info,
			Config: config,
		},
		&reply,
	)
	if err != nil {
		return err
	}
	if len(reply) == 0 {
		return nil
	}
	return errors.New(string(reply))
}

func (c *context) StopService(name string) error {
	return errors.New("not implemented yet")
}

func (c *context) UpdateSystem(configPath string) error {
	err := c.master.Call(
		openedge.CallUpdateSystem,
		&openedge.UpdateSystemRequest{Config: configPath},
		&openedge.UpdateSystemResponse{},
	)
	return err
}

func (c *context) InspectSystem() (*openedge.Inspect, error) {
	reply := &openedge.InspectSystemResponse{}
	err := c.master.Call(
		openedge.CallInspectSystem,
		&openedge.InspectSystemRequest{},
		reply,
	)
	if err != nil {
		return nil, err
	}
	return reply.Inspect, nil
}

func (c *context) ProcessPublish(p *packet.Publish) error {
	if strings.Compare(p.Message.Topic, c.topic) == 0 {
		return c.handle(&openedge.Message{
			Topic:   p.Message.Topic,
			QoS:     byte(p.Message.QOS),
			Payload: p.Message.Payload,
		})
	}
	return nil
}

func (c *context) ProcessPuback(p *packet.Puback) error {
	openedge.Debugln("on puback", p.String())
	return nil
}

func (c *context) ProcessError(err error) {
	openedge.Errorln(err.Error())
}

func newContext() (*context, error) {
	var cfg openedge.Config
	err := utils.LoadYAML(DefaultConfigPath, &cfg)
	if err != nil {
		return nil, err
	}
	err = InitLogger(&cfg.Logger, "service", cfg.Name)
	if err != nil {
		return nil, err
	}

	var jrpccli *jrpc.Client
	addr := os.Getenv(openedge.MasterAPIKey)
	if len(addr) > 0 {
		jrpccli, err = jrpc.NewClient(addr)
		if err != nil {
			return nil, err
		}
	}

	c := &context{
		cfg:    cfg,
		master: jrpccli,
		log:    openedge.GlobalLogger(),
	}
	if len(cfg.Hub.Address) > 0 {
		c.hub = mqtt.NewDispatcher(c.cfg.Hub)
		c.hub.Start(c)
	}
	return c, nil
}

func (c *context) Close() error {
	if c.hub != nil {
		c.hub.Close()
	}
	if c.master != nil {
		c.master.Close()
	}
	return nil
}
