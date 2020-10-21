package chain

import (
	"fmt"
	"io"
	"time"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	v2plugin "github.com/baetyl/baetyl-go/v2/plugin"
	"github.com/baetyl/baetyl-go/v2/pubsub"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"

	"github.com/baetyl/baetyl/v2/ami"
	"github.com/baetyl/baetyl/v2/config"
	"github.com/baetyl/baetyl/v2/plugin"
	"github.com/baetyl/baetyl/v2/sync"
)

const (
	MsgTimeout = time.Minute * 10
)

//go:generate mockgen -destination=../mock/chain.go -package=mock -source=chain.go Chain

type Chain interface {
	Start() error
	io.Closer
}

type chain struct {
	ami       ami.AMI
	data      map[string]string
	upside    string
	downside  string
	pb        plugin.Pubsub
	subChan   chan interface{}
	processor pubsub.Processor
	pipe      ami.Pipe
	tomb      utils.Tomb
	log       *log.Logger
}

func NewChain(cfg config.Config, a ami.AMI, data map[string]string) (Chain, error) {
	pl, err := v2plugin.GetPlugin(cfg.Plugin.Pubsub)
	if err != nil {
		return nil, err
	}

	pipe := ami.Pipe{}
	pipe.InReader, pipe.InWriter = io.Pipe()
	pipe.OutReader, pipe.OutWriter = io.Pipe()

	c := &chain{
		ami:      a,
		data:     data,
		upside:   sync.TopicUpside,
		downside: fmt.Sprintf("%s_%s_%s_%s_%s", data["namespace"], data["name"], data["container"], data["token"], "down"),
		pb:       pl.(plugin.Pubsub),
		pipe:     pipe,
		log:      log.L().With(log.Any("chain", data["token"][:10])),
	}
	c.log.Debug("chain sub downside topic", log.Any("topic", c.downside))
	c.subChan, err = c.pb.Subscribe(c.downside)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *chain) Start() error {
	c.processor = pubsub.NewProcessor(c.subChan, MsgTimeout, &chainHandler{chain: c})
	c.processor.Start()

	return c.tomb.Go(c.debugReading, c.connecting)
}

func (c *chain) Close() error {
	c.processor.Close()

	err := c.pipe.InWriter.Close()
	if err != nil {
		c.log.Warn("failed to close chain in writer", log.Error(err))
	}
	err = c.pipe.OutWriter.Close()
	if err != nil {
		c.log.Warn("failed to close chain out writer", log.Error(err))
	}

	err = c.pb.Unsubscribe(c.downside, c.subChan)
	if err != nil {
		c.log.Warn("failed to unsubscribe chain downside topic", log.Any("topic", c.downside), log.Error(err))
	}
	c.log.Debug("close", log.Any("unsubscribe topic", c.downside))
	return nil
}

func (c *chain) connecting() error {
	cmd := []string{
		"sh",
		"-c",
		"/bin/sh",
	}
	opt := ami.DebugOptions{
		Namespace: c.data["namespace"],
		Name:      c.data["name"],
		Container: c.data["container"],
		Command:   cmd,
	}

	defer func() {
		c.log.Debug("connecting close")
		msg := &v1.Message{
			Kind: v1.MessageCMD,
			Metadata: map[string]string{
				"success": "false",
				"msg":     "disconnect",
				"token":   c.data["token"],
			},
		}
		c.pb.Publish(c.upside, msg)
	}()

	err := c.ami.RemoteCommand(opt, c.pipe)
	if err != nil {
		c.log.Error("failed to start remote debug", log.Error(err))
		return errors.Trace(err)
	}
	return nil
}

func (c *chain) debugReading() error {
	for {
		dt := make([]byte, 10240)
		n, err := c.pipe.OutReader.Read(dt)
		if err != nil && err != io.EOF {
			c.log.Error("failed to read debug message", log.Error(err))
		}
		if err == io.EOF {
			c.log.Info("read debug message EOF")
			return err
		}
		msg := &v1.Message{
			Kind: v1.MessageData,
			Metadata: map[string]string{
				"success": "true",
				"msg":     "ok",
				"token":   c.data["token"],
			},
			Content: v1.LazyValue{Value: dt[0:n]},
		}
		err = c.pb.Publish(c.upside, msg)
		if err != nil {
			c.log.Error("failed to publish message", log.Any("topic", c.upside), log.Error(err))
		}
	}
}
