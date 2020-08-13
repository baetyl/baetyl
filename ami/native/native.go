package native

import (
	"fmt"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl/program"
	"github.com/kardianos/service"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/baetyl/baetyl-go/v2/log"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl/ami"
	"github.com/baetyl/baetyl/config"
)

func init() {
	ami.Register("native", newNativeImpl)
}

type nativeImpl struct {
	log *log.Logger
}

func newNativeImpl(cfg config.AmiConfig) (ami.AMI, error) {
	return &nativeImpl{
		log: log.With(log.Any("ami", "native")),
	}, nil
}

func (impl *nativeImpl) ApplyApp(ns string, app v1.Application, configs map[string]v1.Configuration, secrets map[string]v1.Secret) error {
	appDir := filepath.Join(baetylRunDir, ns, app.Name, app.Version)
	err := os.MkdirAll(appDir, 0755)
	if err != nil {
		return errors.Trace(err)
	}
	avs := map[string]v1.Volume{}
	for _, v := range app.Volumes {
		avs[v.Name] = v
	}
	for _, s := range app.Services {
		for i := 1; i <= s.Replica; i++ {
			// generate instance path
			insDir := filepath.Join(appDir, s.Name, strconv.Itoa(i))
			if err = os.MkdirAll(insDir, 0755); err != nil {
				return errors.Trace(err)
			}

			// apply configuration
			for _, vm := range s.VolumeMounts {
				av, ok := avs[vm.Name]
				if !ok {
					return errors.Errorf("volume (%s) not found in app volumes", vm.Name)
				}
				if av.HostPath != nil {
					os.Symlink(av.HostPath.Path, filepath.Join(insDir, vm.MountPath))
				} else if av.Config != nil {
					vc := configs[av.Config.Name]
					for name, data := range vc.Data {
						file := filepath.Join(insDir, vm.MountPath, name)
						if err = ioutil.WriteFile(file, []byte(data), 0755); err != nil {
							return errors.Trace(err)
						}
					}
				} else if av.Secret != nil {
					vs := secrets[av.Config.Name]
					for name, data := range vs.Data {
						file := filepath.Join(insDir, vm.MountPath, name)
						if err = ioutil.WriteFile(file, data, 0755); err != nil {
							return errors.Trace(err)
						}
					}
				}
			}

			// apply service
			var env []string
			for _, item := range s.Env {
				env = append(env, fmt.Sprintf("%s=%s", item.Name, item.Value))
			}
			prgCfg := program.Config{
				Name:        fmt.Sprintf("%s.%s.%s.%s.%d", ns, app.Name, app.Version, s.Name, i),
				DisplayName: fmt.Sprintf("%s %s", app.Name, s.Name),
				Description: app.Description,
				Dir:         insDir,
				Exec:        s.Image,
				Args:        s.Args,
				Env:         env,
				Logger: log.Config{
					Level:    "debug",
					Filename: filepath.Join(baetylLogDir, app.Name, app.Version, fmt.Sprintf("%s-%d.log", s.Name, i)),
				},
			}
			prgYml, err := yaml.Marshal(prgCfg)
			if err != nil {
				return errors.Trace(err)
			}
			err = ioutil.WriteFile(filepath.Join(insDir, program.DefaultProgramYaml), prgYml, 0755)
			if err != nil {
				return errors.Trace(err)
			}
			svc, err := service.New(nil, &service.Config{
				Name:             prgCfg.Name,
				Description:      prgCfg.Description,
				WorkingDirectory: insDir,
				Arguments:        []string{"program"},
			})
			if err = svc.Install(); err != nil {
				return errors.Trace(err)
			}
		}
	}
	return nil
}

func (impl *nativeImpl) StatsApp(ns string) ([]v1.AppStats, error) {
	return nil, nil
}

func (impl *nativeImpl) DeleteApp(ns string, appName string) error {
	appRootDir := filepath.Join(baetylRunDir, ns, appName)

	// scan app version
	appVerDirs, err := ioutil.ReadDir(appRootDir)
	if err != nil {
		return errors.Trace(err)
	}
	for _, appVerDir := range appVerDirs {
		if !appVerDir.IsDir() {
			continue
		}
		curAppVer := appVerDir.Name()
		// scan service
		svcDirs, err := ioutil.ReadDir(filepath.Join(appRootDir, curAppVer))
		if err != nil {
			return errors.Trace(err)
		}
		for _, svcDir := range svcDirs {
			if !svcDir.IsDir() {
				continue
			}
			curSvcName := svcDir.Name()
			// scan service instance
			svcInsDirs, err := ioutil.ReadDir(filepath.Join(appRootDir, curAppVer, curSvcName))
			if err != nil {
				return errors.Trace(err)
			}
			for _, svcInsDir := range svcInsDirs {
				if !svcInsDir.IsDir() {
					continue
				}
				curSvcIns := svcInsDir.Name()
				prgName := fmt.Sprintf("%s.%s.%s.%s.%s", ns, appName, curAppVer, curSvcName, curSvcIns)
				svc, err := service.New(nil, &service.Config{
					Name:             prgName,
					WorkingDirectory: svcInsDir.Name(),
				})
				if err = svc.Uninstall(); err != nil {
					return errors.Trace(err)
				}
			}
		}
	}
	return nil
}

// TODO: remove
func (impl *nativeImpl) CollectNodeInfo() (*v1.NodeInfo, error) {
	panic("implement me")
}

func (impl *nativeImpl) CollectNodeStats() (*v1.NodeStats, error) {
	panic("implement me")
}

func (impl *nativeImpl) CollectAppStats(s string) ([]v1.AppStats, error) {
	panic("implement me")
}

func (impl *nativeImpl) ApplyApplication(s string, application v1.Application, strings []string) error {
	panic("implement me")
}

func (impl *nativeImpl) ApplyConfigurations(s string, m map[string]v1.Configuration) error {
	panic("implement me")
}

func (impl *nativeImpl) ApplySecrets(s string, m map[string]v1.Secret) error {
	panic("implement me")
}

func (impl *nativeImpl) DeleteApplication(s string, s2 string) error {
	panic("implement me")
}

func (impl *nativeImpl) FetchLog(namespace, service string, tailLines, sinceSeconds int64) (io.ReadCloser, error) {
	panic("implement me")
}
