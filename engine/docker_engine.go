package engine

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/baidu/openedge/config"
	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/module"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/sirupsen/logrus"
)

const defaultNetworkName = "openedge"

// DockerEngine docker engine
type DockerEngine struct {
	context *Context
	client  *client.Client
	network string
	log     *logrus.Entry
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
	pwd, err := filepath.Abs(".")
	if err != nil {
		return nil, err
	}
	exposedPorts, portBindings, err := nat.ParsePortSpecs(m.Expose)
	if err != nil {
		return nil, err
	}
	volumeBindings := []string{
		fmt.Sprintf("%s:/home/openedge/app", path.Join(pwd, "app")),
		fmt.Sprintf("%s:/home/openedge/conf", path.Join(pwd, "app", m.Mark)),
		fmt.Sprintf("%s:/home/openedge/var", path.Join(pwd, "var")),
	}
	cmd := strslice.StrSlice{}
	cmd = append(cmd, m.Params...)
	config := &container.Config{
		Image:        m.Entry,
		ExposedPorts: exposedPorts,
		Cmd:          cmd,
		Env:          module.AppendEnv(m.Env, false),
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
		Spec: Spec{
			Name:    m.Name,
			Restart: m.Restart,
			Grace:   e.context.Grace,
			Logger:  logger.WithFields("module", m.Name),
		},
		Client:        e.client,
		Config:        config,
		HostConfig:    hostConfig,
		NetworkConfig: networkConfig,
	}), err
}

func (e *DockerEngine) initNetwork() error {
	context := context.Background()
	args := filters.NewArgs()
	args.Add("driver", "bridge")
	args.Add("scope", "local")
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
		e.log.Warn(nw.Warning)
	}
	e.network = nw.ID
	e.log.Infof("network (%s:openedge) created", e.network[:12])
	return nil
}
