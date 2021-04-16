package chain

import (
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/pubsub"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"

	"github.com/baetyl/baetyl/v2/ami"
)

func (c *chain) ViewLogs(opt *ami.LogsOptions) error {
	c.logOpt = opt

	c.processor = pubsub.NewProcessor(c.subChan, MsgTimeout, &chainHandler{chain: c})
	c.processor.Start()

	return c.tomb.Go(c.chainReading, c.logging)
}

func (c *chain) logging() error {
	c.logOpt.Name = c.name
	c.logOpt.Namespace = c.namespace
	c.logOpt.Container = c.container

	defer func() {
		c.log.Debug("connecting close")
		msg := &v1.Message{
			Kind: v1.MessageCMD,
			Metadata: map[string]string{
				"success": "false",
				"msg":     "disconnect",
				"token":   c.token,
			},
		}
		c.pb.Publish(c.upside, msg)
	}()

	err := c.ami.RemoteLogs(c.logOpt, c.pipe)
	if err != nil {
		c.log.Error("failed to start view logs", log.Error(err))
		return errors.Trace(err)
	}
	return nil
}
