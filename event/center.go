package event

import (
	"os"
	"sync/atomic"
	"time"

	"github.com/baetyl/baetyl-go/faas"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/utils"
	bh "github.com/timshannon/bolthold"
)

// Handler event handler
type Handler func(faas.Message) error

// Center the event handling center, event handling methods can be registered by topic
type Center struct {
	store    *bh.Store
	limit    int // sets the maximum number of events that can be found by a query
	last     uint64
	signal   chan struct{}
	handlers map[string]Handler
	logger   *log.Logger
	tomb     utils.Tomb
}

// NewCenter create a new persistent center
func NewCenter(store *bh.Store, limit int) (*Center, error) {
	if store == nil || limit < 1 {
		return nil, os.ErrInvalid
	}
	c := &Center{
		store:    store,
		limit:    limit,
		handlers: map[string]Handler{},
		signal:   make(chan struct{}, 1),
		logger:   log.With(log.Any("event", "center")),
	}
	// TODO: to improve bolthold
	last := &faas.Message{}
	num, err := c.store.Count(last, nil)
	if err != nil {
		return nil, err
	}
	if num > 0 {
		err := c.store.FindOne(last, bh.Where(bh.Key).Ge(0).Skip(num-1))
		if err != nil && err != bh.ErrNotFound {
			return nil, err
		}
	}
	c.last = last.ID
	c.Trigger(nil)

	return c, nil
}

// Register register the event handler
func (c *Center) Register(topic string, handler Handler) error {
	if topic == "" || handler == nil {
		return os.ErrInvalid
	}
	c.handlers[topic] = handler
	return nil
}

// Start start the event center
func (c *Center) Start() {
	c.tomb.Go(c.handling)
}

// Close close the event handling center
func (c *Center) Close() error {
	c.tomb.Kill(nil)
	return c.tomb.Wait()
}

// Trigger store event if not nil, then trigger a signal
func (c *Center) Trigger(e *faas.Message) error {
	if e != nil {
		if e.Metadata == nil || e.Metadata["topic"] == "" {
			return os.ErrInvalid
		}
		e.ID = atomic.AddUint64(&c.last, 1)
		err := c.store.Insert(e.ID, e)
		if err != nil {
			return err
		}
		c.logger.Debug("store an event", log.Any("event", e.String()))
	}
	select {
	case c.signal <- struct{}{}:
	default:
	}
	return nil
}

func (c *Center) handling() error {
	c.logger.Info("center starts to handle event")
	defer c.logger.Info("center has stopped handling event")

	var err error
	var events []faas.Message

	for {
		select {
		case <-c.signal:
			events = events[:0]
			err = c.store.Find(&events, bh.Where(bh.Key).Ge(0).Limit(c.limit))
			if err != nil {
				c.logger.Error("failed to find events", log.Error(err))
				time.Sleep(time.Second)
				continue
			}
			if len(events) > 0 {
				c.Trigger(nil) // keep handling events
			}
			// TODO: to merge events if needs
			for _, e := range events {
				c.logger.Debug("find an event", log.Any("event", e.String()))
				topic := e.Metadata["topic"]
				handler, ok := c.handlers[topic]
				if !ok {
					c.logger.Warn("event handler not found", log.Any("event", e.String()))
				}
				err = handler(e)
				if err != nil {
					c.logger.Warn("failed to handle event", log.Error(err), log.Any("event", e.String()))
				}
				err = c.store.Delete(e.ID, &e)
				if err != nil {
					c.logger.Error("failed to delete event", log.Error(err), log.Any("event", e.String()))
				}
			}
		case <-c.tomb.Dying():
			return nil
		}
	}
}
