package main

import (
	"fmt"

	"github.com/baidu/openedge/sdk-go/openedge"
	"github.com/baidu/openedge/utils"
)

// mo bridge module of mqtt servers
type mo struct {
	cfg Config
	rrs []*ruler
}

func main() {
	openedge.Run(func(ctx openedge.Context) error {
		var cfg Config
		err := utils.LoadYAML(openedge.DefaultConfFile, &cfg)
		if err != nil {
			return err
		}
		functions := make(map[string]FunctionInfo)
		for _, f := range cfg.Functions {
			functions[f.Name] = f
		}
		rulers := make([]*ruler, 0)
		for _, ri := range cfg.Rules {
			fi, ok := functions[ri.Function.Name]
			if !ok {
				return fmt.Errorf("function (%s) not found", ri.Function.Name)
			}
			rulers = append(rulers, create(ctx, ri, fi))
		}
		defer func() {
			for _, ruler := range rulers {
				ruler.close()
			}
		}()
		for _, ruler := range rulers {
			err := ruler.start()
			if err != nil {
				return err
			}
		}
		ctx.Wait()
		return nil
	})
}
