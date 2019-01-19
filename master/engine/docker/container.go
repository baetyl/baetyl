package engine

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"time"

	openedge "github.com/baidu/openedge/api/go"
	"github.com/baidu/openedge/master/engine"
	"github.com/baidu/openedge/utils"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/pkg/sysinfo"
	"github.com/docker/go-connections/nat"
)

const serviceNameTemplate = "openedge-service-%s"

type service struct {
	d      *docker
	id     string
	si     *openedge.ServiceInfo
	cfgdir string
	rmcfg  bool
}

func (s *service) Info() *openedge.ServiceInfo {
	return s.si
}

func (s *service) Instances() []engine.Instance {
	return []engine.Instance{}
}

func (s *service) Scale(replica int, grace time.Duration) error {
	return errors.New("not implemented yet")
}

func (s *service) Stop(grace time.Duration) error {
	if err := s.d.client.ContainerStop(context.Background(), s.id, &grace); err != nil {
		openedge.Errorln("failed to stop container:", err.Error())
	}
	err := s.d.client.ContainerRemove(context.Background(), s.id, types.ContainerRemoveOptions{Force: true})
	if s.rmcfg {
		os.RemoveAll(s.cfgdir)
	}
	return err
}

func (d *docker) Run(name string, si *openedge.ServiceInfo) (engine.Service, error) {
	cdir := path.Join(d.wdir, "var", "db", "openedge", "service", name)
	return d.run(name, si, cdir, false)
}

func (d *docker) RunWithConfig(name string, si *openedge.ServiceInfo, config []byte) (engine.Service, error) {
	cdir := path.Join(d.wdir, "var", "run", "openedge", "service", name)
	err := os.MkdirAll(cdir, 0755)
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(path.Join(cdir, "service.yml"), config, 0644)
	if err != nil {
		os.RemoveAll(cdir)
		return nil, err
	}
	s, err := d.run(name, si, cdir, true)
	if err != nil {
		os.RemoveAll(cdir)
		return nil, err
	}
	return s, nil
}

func (d *docker) run(name string, si *openedge.ServiceInfo, cfgdir string, rmcfg bool) (engine.Service, error) {
	if runtime.GOOS == "linux" && si.Resources.CPU.Cpus > 0 {
		sysInfo := sysinfo.New(true)
		if !sysInfo.CPUCfsPeriod || !sysInfo.CPUCfsQuota {
			d.log.Warnf("configuration 'resources.cpu.cpus' is ignored because host kernel does not support CPU cfs period/quota or the cgroup is not mounted.")
			si.Resources.CPU.Cpus = 0
		}
	}
	logdir := path.Join(d.wdir, "var", "log", "openedge", name)
	err := os.MkdirAll(logdir, 0755)
	if err != nil {
		return nil, err
	}
	volumes := make([]string, 0)
	volumes = append(volumes, fmt.Sprintf("%s:%s:ro", cfgdir, "/etc/openedge"))
	volumes = append(volumes, fmt.Sprintf("%s:%s", logdir, "/var/log/openedge"))
	for _, m := range si.Mounts {
		ro := ""
		if m.ReadOnly {
			ro = ":ro"
		}
		volumes = append(volumes, fmt.Sprintf(
			"%s:/%s%s",
			path.Join(d.wdir, "var", "db", "openedge", "service", m.Volume),
			m.Target,
			ro,
		))
	}
	exposedPorts, portBindings, err := nat.ParsePortSpecs(si.Expose)
	cccb, err := d.client.ContainerCreate(
		context.Background(),
		&container.Config{
			Image:        si.Image,
			Env:          utils.AppendEnv(si.Env, false),
			ExposedPorts: exposedPorts,
		},
		&container.HostConfig{
			Binds:        volumes,
			PortBindings: portBindings,
			Resources: container.Resources{
				CpusetCpus: si.Resources.CPU.SetCPUs,
				NanoCPUs:   int64(si.Resources.CPU.Cpus * 1e9),
				Memory:     si.Resources.Memory.Limit,
				MemorySwap: si.Resources.Memory.Swap,
				PidsLimit:  si.Resources.Pids.Limit,
			},
		},
		&network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				defaultNetworkName: &network.EndpointSettings{
					NetworkID: d.network,
				},
			},
		},
		fmt.Sprintf(serviceNameTemplate, name),
	)
	if err != nil {
		return nil, err
	}
	err = d.client.ContainerStart(context.Background(), cccb.ID, types.ContainerStartOptions{})
	if err != nil {
		d.client.ContainerRemove(context.Background(), cccb.ID, types.ContainerRemoveOptions{
			Force: true,
		})
		return nil, err
	}
	return &service{
		d:      d,
		id:     cccb.ID,
		si:     si,
		cfgdir: cfgdir,
		rmcfg:  rmcfg,
	}, nil
}
