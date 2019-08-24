package main

import (
	"context"
	"fmt"

	"github.com/baetyl/baetyl/protocol/mqtt"
	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
)

func main() {
	baetyl.Run(func(ctx baetyl.Context) error {
		var cfg Config
		err := ctx.LoadConfig(&cfg)
		if err != nil {
			return err
		}
		functions := make(map[string]*Function)
		for _, fi := range cfg.Functions {
			functions[fi.Name] = NewFunction(fi, newProducer(ctx, fi))
		}
		defer func() {
			for _, f := range functions {
				f.Close()
			}
		}()
		rulers := make([]*ruler, 0)
		defer func() {
			for _, ruler := range rulers {
				ruler.close()
			}
		}()
		for _, ri := range cfg.Rules {
			f, ok := functions[ri.Function.Name]
			if !ok {
				return fmt.Errorf("function (%s) not found", f.cfg.Name)
			}
			c, err := ctx.NewHubClient(ri.ClientID, []mqtt.TopicInfo{ri.Subscribe})
			if err != nil {
				return fmt.Errorf("failed to create hub client: %s", err.Error())
			}
			rulers = append(rulers, newRuler(ri, c, f))
		}
		for _, ruler := range rulers {
			err := ruler.start()
			if err != nil {
				return err
			}
		}
		if cfg.Server.Address != "" {
			svr, err := baetyl.NewFServer(cfg.Server, func(ctx context.Context, msg *baetyl.FunctionMessage) (*baetyl.FunctionMessage, error) {
				f, ok := functions[msg.FunctionName]
				if !ok {
					return nil, fmt.Errorf("function (%s) not found", msg.FunctionName)
				}
				return f.Call(msg)
			})
			if err != nil {
				return err
			}
			defer svr.Close()
		}
		ctx.Wait()
		return nil
	})
}
