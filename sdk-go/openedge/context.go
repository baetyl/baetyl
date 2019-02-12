package openedge

import (
	"errors"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/256dpi/gomqtt/packet"
	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/protocol/mqtt"
	"github.com/baidu/openedge/utils"
)

// Enviroment variable keys
const (
	EnvHostOSKey       = "OPENEDGE_HOST_OS"
	EnvMasterAPIKey    = "OPENEDGE_MASTER_API"
	EnvServiceModeKey  = "OPENEDGE_SERVICE_MODE"
	EnvServiceNameKey  = "OPENEDGE_SERVICE_NAME"
	EnvServiceTokenKey = "OPENEDGE_SERVICE_TOKEN"
)

// DefaultConfigPath is the path to config of this service
const DefaultConfigPath = "etc/openedge/service.yml"

// Context of module
type Context interface {
	Config() *Config
	// Subscribe(topic mqtt.TopicInfo, handle func(*Message) error) error
	// SendMessage(message *Message) error
	UpdateSystem(*DatasetInfo) error
	InspectSystem() (*Inspect, error)
	Log() logger.Logger
	Wait()
}

type context struct {
	*Client
	cfg    Config
	topic  string
	handle func(*Message) error
	hub    *mqtt.Dispatcher
	log    logger.Logger
}

func (c *context) Config() *Config {
	return &c.cfg
}

func (c *context) Log() logger.Logger {
	return c.log
}

func (c *context) Wait() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	signal.Ignore(syscall.SIGPIPE)
	<-sig
}

func (c *context) Subscribe(topic mqtt.TopicInfo, handle func(*Message) error) error {
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

func (c *context) SendMessage(message *Message) error {
	if c.hub == nil {
		return errors.New("no hub")
	}
	p := packet.NewPublish()
	p.Message.Topic = message.Topic
	p.Message.QOS = packet.QOS(message.QoS)
	p.Message.Payload = message.Payload
	return c.hub.Send(p)
}

func (c *context) ProcessPublish(p *packet.Publish) error {
	if strings.Compare(p.Message.Topic, c.topic) == 0 {
		return c.handle(&Message{
			Topic:   p.Message.Topic,
			QoS:     byte(p.Message.QOS),
			Payload: p.Message.Payload,
		})
	}
	return nil
}

func (c *context) ProcessPuback(p *packet.Puback) error {
	c.log.Debugln("on puback", p.String())
	return nil
}

func (c *context) ProcessError(err error) {
	c.log.Errorln(err.Error())
}

func newContext() (*context, error) {
	var cfg Config
	err := utils.LoadYAML(DefaultConfigPath, &cfg)
	if err != nil {
		return nil, err
	}
	log, err := logger.InitLogger(&cfg.Logger, "service", cfg.Name)
	if err != nil {
		return nil, err
	}

	c := &context{
		cfg: cfg,
		log: log,
	}
	c.Client, err = NewEnvClient()
	if err != nil {
		log.Warnln(err.Error())
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
	return nil
}
