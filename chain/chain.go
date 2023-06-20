package chain

import (
	"context"
	"fmt"
	"io"
	"time"

	v2context "github.com/baetyl/baetyl-go/v2/context"
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
	utils2 "github.com/baetyl/baetyl/v2/utils"
)

const (
	MsgTimeout = time.Minute * 10
	Localhost  = "127.0.0.1"

	ExitCmd = "exit\n"
)

//go:generate mockgen -destination=../mock/chain.go -package=mock -source=chain.go Chain

type Chain interface {
	Debug() error
	ViewLogs(*ami.LogsOptions) error
	Cancel() error
	io.Closer
}

var (
	ErrParseData = errors.New("failed to parse data")
)

type chain struct {
	ami          ami.AMI
	data         map[string]string
	token        string
	upside       string
	downside     string
	mode         string
	debugOptions *ami.DebugOptions
	logOpt       *ami.LogsOptions
	pb           plugin.Pubsub
	subChan      <-chan interface{}
	processor    pubsub.Processor
	pipe         ami.Pipe
	tomb         utils.Tomb
	log          *log.Logger
	ctx          context.Context
	cancel       context.CancelFunc
}

func NewChain(cfg config.Config, a ami.AMI, data map[string]string, needNativeOptions bool) (Chain, error) {
	pl, err := v2plugin.GetPlugin(cfg.Plugin.Pubsub)
	if err != nil {
		return nil, errors.Trace(err)
	}

	pipe := ami.Pipe{}
	pipe.InReader, pipe.InWriter = io.Pipe()
	pipe.OutReader, pipe.OutWriter = io.Pipe()

	ctx, cancel := context.WithCancel(context.Background())
	pipe.Ctx = ctx
	pipe.Cancel = cancel
	c := &chain{
		ami:    a,
		data:   data,
		upside: sync.TopicUpside,
		pb:     pl.(plugin.Pubsub),
		pipe:   pipe,
		ctx:    ctx,
		cancel: cancel,
	}

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

	c.mode = v2context.RunMode()
	cmd := []string{
		"sh",
		"-c",
		"/bin/sh",
	}
	var opt ami.DebugOptions
	// default set kube debug option
	opt.KubeDebugOptions = ami.KubeDebugOptions{
		Namespace: namespace,
		Name:      name,
		Container: container,
		Command:   cmd,
	}
	c.log.Debug("link info", log.Any("data:", data))
	// if host is specified, this is a websocket link. if native mode, use ssh.
	if address, ok := data["host"]; ok {
		path, ok := data["path"]
		if !ok {
			return nil, ErrParseData
		}
		opt.WebsocketOptions = ami.WebsocketOptions{
			Host: address,
			Path: path,
		}
	} else if c.mode == v2context.RunModeNative && true == needNativeOptions {
		port, ok := data["port"]
		if !ok {
			return nil, ErrParseData
		}
		userName, ok := data["userName"]
		if !ok {
			return nil, ErrParseData
		}
		password, ok := data["password"]
		if !ok {
			return nil, ErrParseData
		}
		opt.NativeDebugOptions = ami.NativeDebugOptions{
			IP:       Localhost,
			Port:     port,
			Username: userName,
			Password: password,
		}

	}
	c.debugOptions = &opt
	c.downside = fmt.Sprintf("%s_%s_%s_%s_%s", namespace, name, container, token, "down")

	c.log.Debug("chain sub downside topic", log.Any("topic", c.downside))
	c.subChan, err = c.pb.Subscribe(c.downside)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return c, nil
}

// Cancel Stop ssh with exit cmd and stop websocket with ctx.cancel()
func (c *chain) Cancel() error {
	c.log.Info("connection cancel", log.Any("options", c.debugOptions))
	if c.debugOptions.WebsocketOptions.Host != "" {
		c.cancel()
		return nil
	}
	closeTimer := time.NewTimer(time.Second * 1)
	defer closeTimer.Stop()
	go func() {
		_, err := c.pipe.InWriter.Write([]byte(ExitCmd))
		if err != nil {
			c.log.Error("failed to write debug command", log.Error(err))
		}
		closeTimer.Reset(0)
	}()
	<-closeTimer.C
	return nil
}

func (c *chain) Close() error {
	c.processor.Close()
	c.cancel()
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

func (c *chain) chainReading() error {
	for {
		dt := make([]byte, utils2.ReadBuff)
		n, err := c.pipe.OutReader.Read(dt)
		if err != nil {
			c.log.Error("read remote message close", log.Error(err))
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
		if n >= utils2.KiByte {
			c.log.Debug("ws pipe large data read", log.Any("n", n))
		}
		err = c.pb.Publish(c.upside, msg)
		if err != nil {
			c.log.Error("failed to publish message", log.Any("topic", c.upside), log.Error(err))
		}
	}
}
