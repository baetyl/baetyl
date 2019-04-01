package main

import (
	"encoding/json"
	"time"

	"github.com/baidu/openedge/protocol/mqtt"
	openedge "github.com/baidu/openedge/sdk/openedge-go"
)

type config struct {
	Timer struct {
		Interval time.Duration `yaml:"interval" json:"interval" default:"1m"`
	} `yaml:"timer" json:"timer"`
	Publish mqtt.TopicInfo `yaml:"publish" json:"publish" default:"{\"topic\":\"timer\"}"`
}

func main() {
	// Running service in openedge context
	openedge.Run(func(ctx openedge.Context) error {
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
				msg := map[string]int64{"time": t.Unix()}
				pld, _ := json.Marshal(msg)
				// send a message to hub triggered by timer
				err := cli.Publish(cfg.Publish, pld)
				if err != nil {
					// log error message
					ctx.Log().Errorf(err.Error())
				}
			case <-ctx.WaitChan():
				// wait until service is stopped
				return nil
			}
		}
	})
}
