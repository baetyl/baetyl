package chain

import (
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/pubsub"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"

	"github.com/baetyl/baetyl/v2/ami"
)

func (c *chain) ViewLogs() error {
	c.processor = pubsub.NewProcessor(c.subChan, MsgTimeout, &chainHandler{chain: c})
	c.processor.Start()

	return c.tomb.Go(c.chainReading, c.logging)
}

func (c *chain) logging() error {
	since, ok := c.data["sinceSeconds"].(float64)
	if !ok {
		c.log.Warn(ErrParseData.Error(), log.Any("sinceSeconds", c.data["sinceSeconds"]))
	}
	sinceSeconds := int64(since)
	tail, ok := c.data["tailLines"].(float64)
	if !ok {
		c.log.Warn(ErrParseData.Error(), log.Any("tailLines", c.data["tailLines"]))
	}
	tailLines := int64(tail)
	limit, ok := c.data["limitBytes"].(float64)
	if !ok {
		c.log.Warn(ErrParseData.Error(), log.Any("limitBytes", c.data["limitBytes"]))
	}
	limitBytes := int64(limit)
	follow, ok := c.data["follow"].(bool)
	if !ok {
		c.log.Warn(ErrParseData.Error(), log.Any("follow", c.data["follow"]))
	}
	previous, ok := c.data["previous"].(bool)
	if !ok {
		c.log.Warn(ErrParseData.Error(), log.Any("previous", c.data["previous"]))
	}
	timestamps, ok := c.data["timestamps"].(bool)
	if !ok {
		c.log.Warn(ErrParseData.Error(), log.Any("timestamps", c.data["timestamps"]))
	}

	opt := ami.LogsOptions{
		Namespace:    c.namespace,
		Name:         c.name,
		Container:    c.container,
		SinceSeconds: &sinceSeconds,
		TailLines:    &tailLines,
		LimitBytes:   &limitBytes,
		Follow:       follow,
		Previous:     previous,
		Timestamps:   timestamps,
	}

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

	err := c.ami.RemoteLogs(opt, c.pipe)
	if err != nil {
		c.log.Error("failed to start view logs", log.Error(err))
		return errors.Trace(err)
	}
	return nil
}
