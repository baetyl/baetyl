package main

import (
	"context"
	"fmt"

	"github.com/baidu/openedge/protocol/mqtt"
	openedge "github.com/baidu/openedge/sdk/openedge-go"
)

func main() {
	openedge.Run(func(ctx openedge.Context) error {
		var cfg Config
		err := ctx.LoadConfig(&cfg)
		if err != nil {
			return err
		}
		functions := make(map[string]*Function)
		for _, fi := range cfg.Functions {
			functions[fi.Name] = NewFunction(fi, newProducer(ctx, fi))
		}
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
			svr, err := openedge.NewFServer(cfg.Server, func(ctx context.Context, msg *openedge.FunctionMessage) (*openedge.FunctionMessage, error) {
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
