package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"text/template"
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
		QOS     uint32 `yaml:"qos" json:"qos" validate:"min=0, max=1"`
		Topic   string `yaml:"topic" json:"topic" default:"timer" validate:"nonzero"`
		Payload string `yaml:"payload" json:"payload" default:"{}"`
	} `yaml:"publish" json:"publish"`
}

type cx struct {
	Time  Time
	Rand  Rand
	Frand Frand
}

type Time struct {
}

type Rand struct {
}

type Frand struct {
}

func (t *Time) TE() int64 {
	return time.Now().UnixNano()
}

func (r *Rand) RD() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(100)
}

func (f *Frand) FD() float32 {
	rand.Seed(time.Now().UnixNano())
	return 60 * rand.Float32()
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
			case <-ticker.C:
				//cfg.Publish.Payload["time"] = t.Unix()
				payload := cfg.Publish.Payload
				temp := template.New("timer")
				temp, err := temp.Parse(payload)
				if err != nil {
					return err
				}
				//assert.NoError(ctx, err)
				kk := new(bytes.Buffer)
				err = temp.Execute(kk, &cx{})
				if err != nil {
					return err
				}
				ss := kk.String()
				pkt := packet.NewPublish()
				pkt.Message.Topic = cfg.Publish.Topic
				pkt.Message.QOS = packet.QOS(cfg.Publish.QOS)
				pkt.Message.Payload = []byte(ss)
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
