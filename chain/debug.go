package chain

import (
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/pubsub"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"

	"github.com/baetyl/baetyl/v2/ami"
)

func (c *chain) Debug() error {
	c.processor = pubsub.NewProcessor(c.subChan, MsgTimeout, &chainHandler{chain: c})
	c.processor.Start()

	return c.tomb.Go(c.chainReading, c.connecting)
}

func (c *chain) connecting() error {
	cmd := []string{
		"sh",
		"-c",
		"/bin/sh",
	}
	opt := ami.DebugOptions{
		Namespace: c.namespace,
		Name:      c.name,
		Container: c.container,
		Command:   cmd,
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

	err := c.ami.RemoteCommand(opt, c.pipe)
	if err != nil {
		c.log.Error("failed to start remote debug", log.Error(err))
		return errors.Trace(err)
	}
	return nil
}
