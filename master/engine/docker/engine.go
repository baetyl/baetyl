package docker

import (
	"fmt"
	"path"
	"time"

	"github.com/orcaman/concurrent-map"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/master/engine"
	"github.com/baidu/openedge/utils"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// NAME ot docker engine
const NAME = "docker"

func init() {
	engine.Factories()[NAME] = New
}

// New docker engine
func New(grace time.Duration, pwd string) (engine.Engine, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	e := &dockerEngine{
		cli:   cli,
		pwd:   pwd,
		grace: grace,
		log:   logger.WithField("mode", "docker"),
	}
	err = e.initNetwork()
	if err != nil {
		e.Close()
		return nil, err
	}
	return e, nil
}

type dockerEngine struct {
	cli   *client.Client
	nid   string // network id
	pwd   string // work directory
	grace time.Duration
	log   logger.Logger
}

func (e *dockerEngine) Name() string {
	return NAME
}

func (e *dockerEngine) Close() error {
	return e.cli.Close()
}

// Run a new service
func (e *dockerEngine) Run(cfg engine.ServiceInfo) (engine.Service, error) {
	datasetDir := path.Join(e.pwd, "var", "db", "openedge", "datasets")
	volumes := make([]string, 0)
	for _, m := range cfg.Datasets {
		f := fmtVolume
		if m.ReadOnly {
			f = fmtVolumeRO
		}
		volumes = append(volumes, fmt.Sprintf(f, path.Join(datasetDir, m.Name, m.Version), m.Target))
	}
	for _, m := range cfg.Volumes {
		f := fmtVolume
		if m.ReadOnly {
			f = fmtVolumeRO
		}
		volumes = append(volumes, fmt.Sprintf(f, path.Join(e.pwd, m.Volume), m.Target))
	}
	exposedPorts, portBindings, err := nat.ParsePortSpecs(cfg.Expose)
	if err != nil {
		return nil, err
	}
	cmd := strslice.StrSlice{}
	cmd = append(cmd, cfg.Params...)
	var cfgs containerConfigs
	cfgs.config = &container.Config{
		Image:        cfg.Image,
		Cmd:          cmd,
		Env:          utils.AppendEnv(cfg.Env, false),
		ExposedPorts: exposedPorts,
	}
	cfgs.hostConfig = &container.HostConfig{
		Binds:        volumes,
		PortBindings: portBindings,
		RestartPolicy: container.RestartPolicy{
			Name:              cfg.Restart.Policy,
			MaximumRetryCount: cfg.Restart.Retry.Max,
		},
		Resources: container.Resources{
			CpusetCpus: cfg.Resources.CPU.SetCPUs,
			NanoCPUs:   int64(cfg.Resources.CPU.Cpus * 1e9),
			Memory:     cfg.Resources.Memory.Limit,
			MemorySwap: cfg.Resources.Memory.Swap,
			PidsLimit:  cfg.Resources.Pids.Limit,
		},
	}
	cfgs.networkConfig = &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			defaultNetworkName: &network.EndpointSettings{
				NetworkID: e.nid,
			},
		},
	}
	s := &dockerService{
		info:      cfg,
		engine:    e,
		cfgs:      cfgs,
		instances: cmap.New(),
		log:       e.log.WithField("service", cfg.Name),
	}
	err = s.start()
	if err != nil {
		s.Stop()
		return nil, err
	}
	return s, nil
}
