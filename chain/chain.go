package chain

import (
	"fmt"
	"io"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/mq"
	goplugin "github.com/baetyl/baetyl-go/v2/plugin"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"

	"github.com/baetyl/baetyl/v2/ami"
	"github.com/baetyl/baetyl/v2/config"
	"github.com/baetyl/baetyl/v2/plugin"
)

type Chain interface {
	Publish(interface{}) error
	Subscribe(mq.MQHandler)
	Unsubscribe()
	io.Closer
}

type chainImpl struct {
	data     map[string]string
	upside   string
	downside string
	pipe     *pipe
	ami      ami.AMI
	mq       plugin.MessageQueue
	tomb     utils.Tomb
	log      *log.Logger
}

type pipe struct {
	inReader  io.Reader
	inWriter  io.Writer
	outReader io.Reader
	outWriter io.Writer
}

func NewChain(cfg config.Config, ami ami.AMI, data map[string]string) (Chain, error) {
	mq, err := goplugin.GetPlugin(cfg.Plugin.MQ)
	if err != nil {
		return nil, errors.Trace(err)
	}

	pipe := &pipe{}
	pipe.inReader, pipe.inWriter = io.Pipe()
	pipe.outReader, pipe.outWriter = io.Pipe()

	topic := fmt.Sprintf("%s_%s_%s", data["namespace"], data["name"], data["container"])

	c := &chainImpl{
		data:     data,
		upside:   fmt.Sprintf("%s_%s", topic, "up"),
		downside: fmt.Sprintf("%s_%s", topic, "down"),
		pipe:     pipe,
		ami:      ami,
		mq:       mq.(plugin.MessageQueue),
		log:      log.With(log.Any("chain", topic)),
	}
	c.mq.Subscribe(c.downside, &handlerDownside{chainImpl: c})
	return c, nil
}

func (c *chainImpl) Publish(msg interface{}) error {
	return c.mq.Publish(c.downside, msg)
}

func (c *chainImpl) Subscribe(handler mq.MQHandler) {
	c.mq.Subscribe(c.upside, handler)
}

func (c *chainImpl) Unsubscribe() {
	c.mq.Unsubscribe(c.upside)
}

func (c *chainImpl) Close() error {
	c.tomb.Kill(nil)
	c.tomb.Wait()
	c.mq.Unsubscribe(c.downside)
	return nil
}

func (c *chainImpl) connecting() error {
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
		msg := &v1.Message{
			Kind: v1.MessageCMD,
			Metadata: map[string]string{
				"success": "false",
				"msg":     "disconnect",
			},
		}
		c.mq.Publish(c.upside, msg)
	}()

	c.tomb.Go(c.debugReading)

	err := c.ami.RemoteCommand(opt, c.pipe.inReader, c.pipe.outWriter, c.pipe.outWriter)
	if err != nil {
		c.log.Error("failed to start remote debug", log.Error(err))
		return errors.Trace(err)
	}
	return nil
}

func (c *chainImpl) debugReading() error {
	for {
		dt := make([]byte, 10240)
		n, err := c.pipe.outReader.Read(dt)
		if err != nil && err != io.EOF {
			c.log.Error("failed to read debug message")
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
			},
			Content: v1.LazyValue{Value: dt[0:n]},
		}
		err = c.mq.Publish(c.upside, msg)
		if err != nil {
			c.log.Error("failed to publish message")
		}
	}
}
