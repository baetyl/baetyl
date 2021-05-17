package chain

import (
	"fmt"
	"io"
	"time"

	"github.com/baetyl/baetyl-go/v2/context"
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
	Debug() error
	ViewLogs(*ami.LogsOptions) error
	io.Closer
}

var (
	ErrParseData = errors.New("failed to parse data")
)

type chain struct {
	mode         string
	ami          ami.AMI
	data         map[string]string
	token        string
	debugOption  ami.DebugOptions
	upside       string
	downside     string
	logOpt       *ami.LogsOptions
	pb           plugin.Pubsub
	subChan      <-chan interface{}
	processor    pubsub.Processor
	debugSession io.Closer
	pipe         ami.Pipe
	tomb         utils.Tomb
	log          *log.Logger
}

func NewChain(cfg config.Config, a ami.AMI, data map[string]string) (Chain, error) {
	pl, err := v2plugin.GetPlugin(cfg.Plugin.Pubsub)
	if err != nil {
		return nil, errors.Trace(err)
	}

	pipe := ami.Pipe{}
	pipe.InReader, pipe.InWriter = io.Pipe()
	pipe.OutReader, pipe.OutWriter = io.Pipe()

	c := &chain{
		mode:   context.RunMode(),
		ami:    a,
		data:   data,
		upside: sync.TopicUpside,
		pb:     pl.(plugin.Pubsub),
		pipe:   pipe,
	}

	mode := context.RunMode()
	platform := context.PlatformString()
	for k, _ := range cfg.Data {
		if !specv1.IsConfigObject(k) {
			continue
		}
		if mode == context.RunModeNative && strings.Contains(k, platform) {
			continue
		}
		delete(cfg.Data, k)
	}

	if mode == context.RunModeNative {
		port, ok := data["port"]
		if !ok {
			c.log.Debug("no port specified")
		}
		username, ok := data["username"]
		if !ok {
			c.log.Debug("no username specified")
		}
		password, ok := data["password"]
		if !ok {
			c.log.Debug("no password specified")
		}

		c.debugOption.UserName = username
		c.debugOption.Password = password
		c.debugOption.Port = port
	} else {
		token, ok := data["token"]
		if !ok {
			return nil, ErrParseData
		}
		c.token = token
		c.log = log.L().With(log.Any("chain", token))

		name, ok := data["name"]
		if !ok {
			return nil, ErrParseData
		}
		namespace, ok := data["namespace"]
		if !ok {
			return nil, ErrParseData
		}
		container, ok := data["container"]
		if !ok {
			c.log.Debug("no container specified")
		}
		c.debugOption.Name = name
		c.debugOption.Namespace = namespace
		c.debugOption.Container = container
	}

	// TODO: can't user container here
	c.downside = fmt.Sprintf("%s_%s_%s_%s_%s", namespace, name, container, token, "down")

	c.log.Debug("chain sub downside topic", log.Any("topic", c.downside))
	c.subChan, err = c.pb.Subscribe(c.downside)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return c, nil
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
	err = c.debugSession.Close()
	if err != nil {
		c.log.Warn("failed to close debug session of chain", log.Error(err))
	}

	err = c.pb.Unsubscribe(c.downside, c.subChan)
	if err != nil {
		c.log.Warn("failed to unsubscribe chain downside topic", log.Any("topic", c.downside), log.Error(err))
	}
	c.log.Debug("close", log.Any("unsubscribe topic", c.downside))
	return nil
}

func (c *chain) chainReading() error {
	for {
		dt := make([]byte, 10240)
		n, err := c.pipe.OutReader.Read(dt)
		if err != nil && err != io.EOF {
			c.log.Error("failed to read remote message", log.Error(err))
		}
		if err == io.EOF {
			c.log.Info("read remote message EOF")
			return errors.Trace(err)
		}
		msg := &v1.Message{
			Kind: v1.MessageData,
			Metadata: map[string]string{
				"success": "true",
				"msg":     "ok",
				"token":   c.token,
			},
			Content: v1.LazyValue{Value: dt[0:n]},
		}
		err = c.pb.Publish(c.upside, msg)
		if err != nil {
			c.log.Error("failed to publish message", log.Any("topic", c.upside), log.Error(err))
		}
	}
}
