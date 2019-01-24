package engine

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"

	openedge "github.com/baidu/openedge/api/go"
	"github.com/baidu/openedge/master/engine"
	"github.com/baidu/openedge/utils"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/sysinfo"
	"github.com/docker/go-connections/nat"
)

// NAME ot docker engine
const NAME = "docker"
const defaultNetworkName = "openedge"

func init() {
	engine.Factories()[NAME] = New
}

// New docker engine
func New(wdir string) (engine.Engine, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	e := &dockerEngine{
		client: cli,
		wdir:   wdir,
		log:    openedge.WithField("mode", "docker"),
	}
	err = e.initNetwork()
	if err != nil {
		return nil, err
	}
	return e, nil
}

type dockerEngine struct {
	client  *client.Client
	wdir    string
	network string
	log     openedge.Logger
}

func (e *dockerEngine) Name() string {
	return NAME
}

func (e *dockerEngine) Close() error {
	return e.client.Close()
}

func (e *dockerEngine) initNetwork() error {
	context := context.Background()
	args := filters.NewArgs()
	args.Add("driver", "bridge")
	args.Add("type", "custom")
	args.Add("name", defaultNetworkName)
	nws, err := e.client.NetworkList(context, types.NetworkListOptions{Filters: args})
	if err != nil {
		return err
	}
	if len(nws) > 0 {
		e.network = nws[0].ID
		e.log.Infof("network (%s:openedge) exists", e.network[:12])
		return nil
	}
	nw, err := e.client.NetworkCreate(context, defaultNetworkName, types.NetworkCreate{Driver: "bridge", Scope: "local"})
	if err != nil {
		return err
	}
	if nw.Warning != "" {
		e.log.Warnf(nw.Warning)
	}
	e.network = nw.ID
	e.log.Infof("network (%s:openedge) created", e.network[:12])
	return nil
}

// Run new service
func (e *dockerEngine) Run(si *openedge.ServiceInfo) (engine.Service, error) {
	cdir := path.Join(e.wdir, "var", "db", "openedge", "service", si.Name)
	return e.run(si, cdir, false)
}

// RunWithConfig new service
func (e *dockerEngine) RunWithConfig(si *openedge.ServiceInfo, config []byte) (engine.Service, error) {
	cdir := path.Join(e.wdir, "var", "run", "openedge", "service", si.Name)
	err := os.MkdirAll(cdir, 0755)
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(path.Join(cdir, "service.yml"), config, 0644)
	if err != nil {
		os.RemoveAll(cdir)
		return nil, err
	}
	s, err := e.run(si, cdir, true)
	if err != nil {
		os.RemoveAll(cdir)
		return nil, err
	}
	return s, nil
}

func (e *dockerEngine) run(si *openedge.ServiceInfo, cfgdir string, rmcfg bool) (engine.Service, error) {
	if runtime.GOOS == "linux" && si.Resources.CPU.Cpus > 0 {
		sysInfo := sysinfo.New(true)
		if !sysInfo.CPUCfsPeriod || !sysInfo.CPUCfsQuota {
			e.log.Warnf("configuration 'resources.cpu.cpus' is ignored because host kernel does not support CPU cfs period/quota or the cgroup is not mounted.")
			si.Resources.CPU.Cpus = 0
		}
	}
	logdir := path.Join(e.wdir, "var", "log", "openedge", si.Name)
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
			path.Join(e.wdir, m.Volume),
			m.Target,
			ro,
		))
	}
	exposedPorts, portBindings, err := nat.ParsePortSpecs(si.Expose)
	cccb, err := e.client.ContainerCreate(
		context.Background(),
		&container.Config{
			Image:        si.Image,
			Env:          utils.AppendEnv(si.Env, false),
			ExposedPorts: exposedPorts,
		},
		&container.HostConfig{
			Binds:        volumes,
			PortBindings: portBindings,
			RestartPolicy: container.RestartPolicy{
				Name:              si.Restart.Policy,
				MaximumRetryCount: si.Restart.Retry.Max,
			},
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
					NetworkID: e.network,
				},
			},
		},
		si.Name,
	)
	if err != nil {
		return nil, err
	}
	err = e.client.ContainerStart(context.Background(), cccb.ID, types.ContainerStartOptions{})
	if err != nil {
		e.client.ContainerRemove(context.Background(), cccb.ID, types.ContainerRemoveOptions{
			Force: true,
		})
		return nil, err
	}
	return &dockerService{
		e:      e,
		id:     cccb.ID,
		si:     si,
		cfgdir: cfgdir,
		rmcfg:  rmcfg,
	}, nil
}

/*
// Prepare prepares images
func (e *DockerEngine) Prepare(image string) error {
	out, err := e.client.ImagePull(context.Background(), image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer out.Close()
	io.Copy(os.Stdout, out)
	return nil
}
*/
