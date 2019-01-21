package sdk

import (
	"errors"
	"io/ioutil"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/256dpi/gomqtt/packet"
	openedge "github.com/baidu/openedge/api/go"
	"github.com/baidu/openedge/protocol/mqtt"
	"github.com/baidu/openedge/utils"
)

// DefaultConfigPath is the path to config of this service
const DefaultConfigPath = "etc/openedge/service.yml"

type context struct {
	mqtt.Handler
	cfg     openedge.Config
	mqtt    *mqtt.Dispatcher
	topic   string
	handler func(*openedge.Message) error
	master  *rpc.Client
}

func (c *context) Config() *openedge.Config {
	return &c.cfg
}

func (c *context) WaitExit() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	signal.Ignore(syscall.SIGPIPE)
	<-sig
}

func (c *context) Subscribe(topic openedge.TopicInfo, handler func(*openedge.Message) error) error {
	if c.mqtt == nil {
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
	err := c.mqtt.Send(p)
	if err != nil {
		return err
	}
	// FIXME not support multiple subscription
	c.topic = topic.Topic
	c.handler = handler
	return nil
}

func (c *context) SendMessage(message *openedge.Message) error {
	if c.mqtt == nil {
		return errors.New("no hub")
	}
	p := packet.NewPublish()
	p.Message.Topic = message.Topic
	p.Message.QOS = packet.QOS(message.QoS)
	p.Message.Payload = message.Payload
	return c.mqtt.Send(p)
}

func (c *context) StartService(name string, info *openedge.ServiceInfo, config []byte) error {
	var reply openedge.StartServiceResponse
	err := c.master.Call(
		openedge.CallStartService,
		openedge.StartServiceRequest{
			Name:   name,
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
	return errors.New("not implemented yet")
}

func (c *context) onPublish(p *packet.Publish) error {
	if strings.Compare(p.Message.Topic, c.topic) == 0 {
		return c.handler(&openedge.Message{
			Topic:   p.Message.Topic,
			QoS:     byte(p.Message.QOS),
			Payload: p.Message.Payload,
		})
	}
	return nil
}

func (c *context) onPuback(p *packet.Puback) error {
	openedge.Debugln("on puback", p.String())
	return nil
}

func (c *context) onError(err error) {

}

func newContext() (*context, error) {
	c := &context{
		cfg: openedge.Config{},
	}
	data, err := ioutil.ReadFile(DefaultConfigPath)
	if err != nil {
		return nil, err
	}
	err = utils.UnmarshalYAML(data, &c.cfg)
	if err != nil {
		return nil, err
	}
	err = InitLogger(&c.cfg.Logger)
	if err != nil {
		return nil, err
	}
	if len(c.cfg.Hub.Address) > 0 {
		c.mqtt = mqtt.NewDispatcher(c.cfg.Hub)
		c.ProcessPublish = c.onPublish
		c.ProcessPuback = c.onPuback
		c.ProcessError = c.onError
		c.mqtt.Start(c.Handler)
	}

	master := os.Getenv(openedge.MasterAPIKey)
	if len(master) > 0 {
		addr, err := url.Parse(master)
		if err != nil {
			return nil, err
		}
		conn, err := net.Dial(addr.Scheme, addr.Host)
		if err != nil {
			return nil, err
		}
		c.master = jsonrpc.NewClient(conn)
	}
	return c, nil
}

func (c *context) Close() error {
	if c.mqtt != nil {
		c.mqtt.Close()
	}
	if c.master != nil {
		c.master.Close()
	}
	return nil
}
