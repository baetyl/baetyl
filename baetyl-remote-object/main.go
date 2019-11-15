package main

import (
	"fmt"

	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
)

// mo bridge module of mqtt servers
type mo struct {
	cfg Config
	rrs []*Ruler
}

func main() {
	baetyl.Run(func(ctx baetyl.Context) error {
		var cfg Config
		err := ctx.LoadConfig(&cfg)
		if err != nil {
			return err
		}
		// clients
		clients := make(map[string]Client)
		for _, c := range cfg.Clients {
			clients[c.Name], err = NewClient(c, ctx.ReportInstance)
			if err != nil {
				return err
			}
		}
		// rulers
		rulers := make([]*Ruler, 0)
		for _, r := range cfg.Rules {
			cli, ok := clients[r.Client.Name]
			if !ok {
				return fmt.Errorf("client (%s) not found", r.Client.Name)
			}
			ruler := NewRuler(ctx, r, cli)
			rulers = append(rulers, ruler)
		}
		defer func() {
			for _, ruler := range rulers {
				ruler.Close()
			}
			for _, cli := range clients {
				cli.Close()
			}
		}()
		for _, cli := range clients {
			err := cli.Start()
			if err != nil {
				return err
			}
		}
		for _, ruler := range rulers {
			err := ruler.Start()
			if err != nil {
				return err
			}
		}
		ctx.Wait()
		return nil
	})
}
