package sync

import (
	"os"
	"path/filepath"

	"github.com/baetyl/baetyl-go/v2/context"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
)

// TODO: move to the right place
func PrepareApp(hostPath, objectPath string, app *specv1.Application, cfgs map[string]specv1.Configuration) error {
	if app == nil {
		return nil
	}
	for i := 0; i < len(app.Services); i++ {
		env := []specv1.Environment{
			{
				Name:  context.KeyAppName,
				Value: app.Name,
			},
			{
				Name:  context.KeySvcName,
				Value: app.Services[i].Name,
			},
			{
				Name:  context.KeyAppVersion,
				Value: app.Version,
			},
			{
				Name:  context.KeyNodeName,
				Value: os.Getenv(context.KeyNodeName),
			},
			{
				Name:  context.KeyRunMode,
				Value: context.RunMode(),
			},
		}
		app.Services[i].Env = append(app.Services[i].Env, env...)
	}
	for i := range app.Volumes {
		if hp := app.Volumes[i].HostPath; hp != nil {
			if filepath.IsAbs(hp.Path) {
				continue
			}
			fullPath := filepath.Join(hostPath, filepath.Join("/", hp.Path))
			if err := os.MkdirAll(fullPath, 0755); err != nil {
				return err
			}
			app.Volumes[i].HostPath = &specv1.HostPathVolumeSource{Path: fullPath}
		} else if config := app.Volumes[i].Config; config != nil {
			cfg, ok := cfgs[config.Name]
			if !ok {
				continue
			}
			for k := range cfg.Data {
				if !specv1.IsConfigObject(k) {
					continue
				}
				if app.Volumes[i].HostPath == nil {
					app.Volumes[i].Config = nil
					app.Volumes[i].HostPath = &specv1.HostPathVolumeSource{
						Path: filepath.Join(objectPath, cfg.Name, cfg.Version),
					}
				}
			}
		}
	}
	return nil
}
