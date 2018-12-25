package engine

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/baidu/openedge/module/config"
	"github.com/baidu/openedge/module/logger"
	"github.com/baidu/openedge/module/utils"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

const defaultNetworkName = "openedge"

// DockerEngine docker engine
type DockerEngine struct {
	context *Context
	client  *client.Client
	network string
	log     *logger.Entry
}

// NewDockerEngine create a new docker engine
func NewDockerEngine(ctx *Context) (Inner, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	e := &DockerEngine{
		context: ctx,
		client:  cli,
		log:     logger.WithFields("mode", "docker"),
	}
	return e, e.initNetwork()
}

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

// Create creates a new docker container
func (e *DockerEngine) Create(m config.Module) (Worker, error) {
	exposedPorts, portBindings, err := nat.ParsePortSpecs(m.Expose)
	if err != nil {
		return nil, err
	}
	logPath := path.Join(e.context.PWD, "var", "log", "openedge", m.Name)
	volumePath := path.Join(e.context.PWD, "var", "db", "openedge", "volume", m.Name)
	modulePath := path.Join(e.context.PWD, "var", "db", "openedge", "module", m.Name)
	configPath := path.Join(modulePath, "module.yml")
	volumeBindings := []string{
		fmt.Sprintf("%s:/etc/openedge/module.yml:ro", configPath),
		fmt.Sprintf("%s:/var/db/openedge/module/%s:ro", modulePath, m.Name),
		fmt.Sprintf("%s:/var/db/openedge/volume/%s", volumePath, m.Name),
		fmt.Sprintf("%s:/var/log/openedge/%s", logPath, m.Name),
	}
	cmd := strslice.StrSlice{}
	cmd = append(cmd, m.Params...)
	config := &container.Config{
		Image:        m.Entry,
		ExposedPorts: exposedPorts,
		Cmd:          cmd,
		Env:          utils.AppendEnv(m.Env, false),
	}
	hostConfig := &container.HostConfig{
		Binds:        volumeBindings,
		PortBindings: portBindings,
		Resources: container.Resources{
			CpusetCpus: m.Resources.CPU.SetCPUs,
			NanoCPUs:   int64(m.Resources.CPU.Cpus * 1e9),
			Memory:     m.Resources.Memory.Limit,
			MemorySwap: m.Resources.Memory.Swap,
			PidsLimit:  m.Resources.Pids.Limit,
		},
	}
	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			defaultNetworkName: &network.EndpointSettings{
				NetworkID: e.network,
			},
		},
	}
	return NewDockerContainer(&DockerSpec{
		module:        &m,
		context:       e.context,
		client:        e.client,
		config:        config,
		hostConfig:    hostConfig,
		networkConfig: networkConfig,
	}), err
}

func (e *DockerEngine) initNetwork() error {
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
