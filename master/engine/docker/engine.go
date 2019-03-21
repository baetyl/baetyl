package docker

import (
	"fmt"
	"path"
	"runtime"
	"sync"
	"time"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/master/engine"
	openedge "github.com/baidu/openedge/sdk/openedge-go"
	"github.com/baidu/openedge/utils"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/sysinfo"
	"github.com/docker/go-connections/nat"
	"github.com/orcaman/concurrent-map"
	"strings"
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
		log:   logger.WithField("engine", NAME),
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

// Prepare prepares all images
func (e *dockerEngine) Prepare(ss []openedge.ServiceInfo) {
	var wg sync.WaitGroup
	for _, s := range ss {
		wg.Add(1)
		go func(i string, w *sync.WaitGroup) {
			defer w.Done()
			e.pullImage(i)
		}(s.Image, &wg)
	}
	wg.Wait()
}

// Run a new service
func (e *dockerEngine) Run(cfg openedge.ServiceInfo, vs map[string]openedge.VolumeInfo) (engine.Service, error) {

	if runtime.GOOS == "linux" && cfg.Resources.CPU.Cpus > 0 {
		sysInfo := sysinfo.New(true)
		if !sysInfo.CPUCfsPeriod || !sysInfo.CPUCfsQuota {
			e.log.Warnf("configuration 'resources.cpu.cpus' of service (%s) is ignored, because host kernel does not support CPU cfs period/quota or the cgroup is not mounted.", cfg.Name)
			cfg.Resources.CPU.Cpus = 0
		}
	}

	volumes := make([]string, 0)
	for _, m := range cfg.Mounts {
		v, ok := vs[m.Name]
		if !ok {
			return nil, fmt.Errorf("volume '%s' not found", m.Name)
		}
		f := fmtVolume
		if m.ReadOnly {
			f = fmtVolumeRO
		}
		volumes = append(volumes, fmt.Sprintf(f, path.Join(e.pwd, path.Clean(v.Path)), path.Clean(m.Path)))
	}
	exposedPorts, portBindings, err := nat.ParsePortSpecs(cfg.Ports)
	if err != nil {
		return nil, err
	}
	deviceBindings, err := e.parseDeviceSpecs(cfg.Devices)
	if err != nil {
		return nil, err
	}
	var params containerConfigs
	params.config = container.Config{
		Image:        cfg.Image,
		Env:          utils.AppendEnv(cfg.Env, false),
		ExposedPorts: exposedPorts,
	}
	params.hostConfig = container.HostConfig{
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
			Devices:    deviceBindings,
		},
	}
	params.networkConfig = network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			defaultNetworkName: &network.EndpointSettings{
				NetworkID: e.nid,
			},
		},
	}
	s := &dockerService{
		cfg:       cfg,
		engine:    e,
		params:    params,
		instances: cmap.New(),
		log:       e.log.WithField("service", cfg.Name),
	}
	err = s.Start()
	if err != nil {
		s.Stop()
		return nil, err
	}
	return s, nil
}

func (e *dockerEngine) parseDeviceSpecs(devices []string) (deviceBindings []container.DeviceMapping, err error) {
	for _, rawDevice := range devices {
		deviceMapping := strings.Split(rawDevice, ":")
		device := container.DeviceMapping{}
		switch len(deviceMapping) {
		case 3:
			device.CgroupPermissions = deviceMapping[2]
		case 2:
			device.CgroupPermissions = "mrw"
		default:
			err = fmt.Errorf("error device mapping[%s]", rawDevice)
			return
		}
		device.PathOnHost = deviceMapping[0]
		device.PathInContainer = deviceMapping[1]
		deviceBindings = append(deviceBindings, device)
	}
	return
}
