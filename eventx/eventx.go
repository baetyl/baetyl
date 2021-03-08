package eventx

import (
	"io"

	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/mqtt"
	goplugin "github.com/baetyl/baetyl-go/v2/plugin"
	"github.com/baetyl/baetyl-go/v2/pubsub"

	"github.com/baetyl/baetyl/v2/config"
	"github.com/baetyl/baetyl/v2/plugin"
)

const (
	TopicEvent = "event"
)

type EventX interface {
	Start()
	io.Closer
}

type eventX struct {
	pb        plugin.Pubsub
	mqtt      *mqtt.Client
	log       *log.Logger
	processor pubsub.Processor
}

func NewEventX(ctx context.Context, cfg config.Config) (EventX, error) {
	pl, err := goplugin.GetPlugin(cfg.Plugin.Pubsub)
	if err != nil {
		return nil, errors.Trace(err)
	}
	mqtt, err := ctx.NewSystemBrokerClient(nil)
	if err != nil {
		return nil, err
	}
	pb := pl.(plugin.Pubsub)
	ch, err := pb.Subscribe(TopicEvent)
	if err != nil {
		return nil, err
	}
	log := log.With(log.Any("core", "sync"))
	processor := pubsub.NewProcessor(ch, 0, &handler{mqtt: mqtt, log: log, cfg: cfg.Event})
	evt := &eventX{
		pb:        pb,
		mqtt:      mqtt,
		processor: processor,
		log:       log,
	}
	return evt, nil
}

func (e *eventX) Start() {
	if err := e.mqtt.Start(nil); err != nil {
		e.log.Warn("failed to start mqtt client", log.Error(err))
	}
	e.processor.Start()
}

func (e *eventX) Close() error {
	e.processor.Close()
	return e.mqtt.Close()
}
