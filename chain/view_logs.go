package chain

import (
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/pubsub"

	"github.com/baetyl/baetyl/v2/ami"
)

func (c *chain) ViewLogs(opt *ami.LogsOptions) error {
	c.logOpt = opt

	c.processor = pubsub.NewProcessor(c.subChan, 0, &chainHandler{chain: c})
	c.processor.Start()

	return c.tomb.Go(c.chainReading, c.logging)
}

func (c *chain) logging() error {
	c.logOpt.Name = c.debugOptions.Name
	c.logOpt.Namespace = c.debugOptions.Namespace
	c.logOpt.Container = c.debugOptions.Container

	err := c.ami.RemoteLogs(c.logOpt, c.pipe)
	if err != nil {
		c.log.Error("failed to start view logs", log.Error(err))
		return errors.Trace(err)
	}
	return nil
}
