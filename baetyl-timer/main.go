package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/256dpi/gomqtt/packet"
	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
)

// custom configuration of the timer module
type config struct {
	Timer struct {
		Interval time.Duration `yaml:"interval" json:"interval" default:"1m"`
	} `yaml:"timer" json:"timer"`
	Publish struct {
		QOS     uint32                 `yaml:"qos" json:"qos" validate:"min=0, max=1"`
		Topic   string                 `yaml:"topic" json:"topic" default:"timer" validate:"nonzero"`
		Payload map[string]interface{} `yaml:"payload" json:"payload" default:"{}"`
	} `yaml:"publish" json:"publish"`
}

func main() {
	// Running module in baetyl context
	baetyl.Run(func(ctx baetyl.Context) error {
		var cfg config
		// load custom config
		err := ctx.LoadConfig(&cfg)
		if err != nil {
			return err
		}
		// create a hub client
		cli, err := ctx.NewHubClient("", nil)
		if err != nil {
			return err
		}
		// start client to keep connection with hub
		cli.Start(nil)
		// create a timer
		ticker := time.NewTicker(cfg.Timer.Interval)
		defer ticker.Stop()
		for {
			select {
			case t := <-ticker.C:
				cfg.Publish.Payload["time"] = t.Unix()
				pld, err := json.Marshal(cfg.Publish.Payload)
				if err != nil {
					return fmt.Errorf("Failed to marshal: %s", err.Error())
				}
				pkt := packet.NewPublish()
				pkt.Message.Topic = cfg.Publish.Topic
				pkt.Message.QOS = packet.QOS(cfg.Publish.QOS)
				pkt.Message.Payload = pld
				// send a message to hub triggered by timer
				err = cli.Send(pkt)
				if err != nil {
					return fmt.Errorf("Failed to publish: %s", err.Error())
				}
			case <-ctx.WaitChan():
				// wait until service is stopped
				return nil
			}
		}
	})
}
