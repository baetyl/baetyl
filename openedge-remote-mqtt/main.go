package main

import (
	"fmt"

	openedge "github.com/baidu/openedge/api/go"
	sdk "github.com/baidu/openedge/sdk/go"
	"github.com/baidu/openedge/utils"
)

// mo bridge module of mqtt servers
type mo struct {
	cfg Config
	rrs []*ruler
}

const defaultConfigPath = "/etc/openedge/service.yml"

func main() {
	sdk.Run(func(ctx openedge.Context) error {
		var cfg Config
		err := utils.LoadYAML(defaultConfigPath, &cfg)
		if err != nil {
			return err
		}
		remotes := make(map[string]Remote)
		for _, remote := range cfg.Remotes {
			remotes[remote.Name] = remote
		}
		rulers := make([]*ruler, 0)
		for _, rule := range cfg.Rules {
			remote, ok := remotes[rule.Remote.Name]
			if !ok {
				return fmt.Errorf("remote (%s) not found", rule.Remote.Name)
			}
			rulers = append(rulers, create(rule, cfg.Hub, remote.MqttClientInfo))
		}
		defer func() {
			for _, ruler := range rulers {
				ruler.close()
			}
		}()
		for _, ruler := range rulers {
			err := ruler.start()
			if err != nil {
				openedge.WithError(err).Errorf("failed to start rule")
				return err
			}
		}
		ctx.WaitExit()
		return nil
	})
}
