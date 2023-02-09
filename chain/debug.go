package chain

import (
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/pubsub"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
)

func (c *chain) Debug() error {
	c.processor = pubsub.NewProcessor(c.subChan, MsgTimeout, &chainHandler{chain: c})
	c.processor.Start()

	return c.tomb.Go(c.chainReading, c.connecting)
}

func (c *chain) connecting() error {
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

	var err error
	// if host is specified, this is a websocket cmd, else ssh
	if c.debugOptions.WebsocketOptions.Host != "" {
		err = c.ami.RemoteWebsocket(c.ctx, c.debugOptions, c.pipe)
	} else {
		err = c.ami.RemoteCommand(c.debugOptions, c.pipe)
	}
	if err != nil {
		c.log.Error("failed to start remote debug", log.Error(err))
		c.Close()
		return errors.Trace(err)
	}
	return nil
}
