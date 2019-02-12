package main

import (
	"fmt"

	"github.com/baidu/openedge/sdk-go/openedge"
	"github.com/baidu/openedge/utils"
	yaml "gopkg.in/yaml.v2"
)

func run(ctx openedge.Context) error {
	var cfg Config
	err := utils.LoadYAML(openedge.DefaultConfigPath, &cfg)
	if err != nil {
		return err
	}
	for _, f := range cfg.Functions {
		hub := ctx.Config().Hub
		hub.ClientID = fmt.Sprintf("openedge-function-%s", f.Name)
		rt := RuntimeInfo{
			Config: openedge.Config{
				Hub:    hub,
				Logger: ctx.Config().Logger,
			},
			Subscribe: f.Subscribe,
			Publish:   f.Publish,
			Name:      f.Name,
			Handler:   f.Handler,
		}
		_, err := yaml.Marshal(&rt)
		if err != nil {
			return err
		}
		// TODO: fix
		// si := openedge.ServiceInfo{
		// 	Name:    f.Name,
		// 	Image:   f.Runtime,
		// 	Replica: 1,
		// 	Expose:  []string{},
		// 	Params:  []string{},
		// 	Env:     f.Env,
		// 	Mounts: []openedge.MountInfo{
		// 		openedge.MountInfo{
		// 			Volume: f.CodeDir,
		// 			Target: "code",
		// 			// ReadOnly: true,
		// 		},
		// 	},
		// 	// TODO Restart
		// 	// TODO Resource
		// }
		// err = ctx.StartService(&si, rtcfg)
		// if err != nil {
		// 	return err
		// }
	}
	ctx.Wait()
	return nil
}

func main() {
	openedge.Run(run)
}
