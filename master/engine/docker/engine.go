package docker

import (
	"fmt"
	"path"
	"runtime"
	"strings"
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
	cmap "github.com/orcaman/concurrent-map"
)

// NAME ot docker engine
const NAME = "docker"

func init() {
	engine.Factories()[NAME] = New
}

// New docker engine
func New(grace time.Duration, pwd string, stats engine.InfoStats) (engine.Engine, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	e := &dockerEngine{
		InfoStats: stats,
		cli:       cli,
		pwd:       pwd,
		grace:     grace,
		log:       logger.WithField("engine", NAME),
	}
	err = e.initNetwork()
	if err != nil {
		e.Close()
		return nil, err
	}
	return e, nil
}

type dockerEngine struct {
	engine.InfoStats
	cli   *client.Client
	nid   string // network id
	pwd   string // work directory
	grace time.Duration
	tomb  utils.Tomb
	log   logger.Logger
}

func (e *dockerEngine) Name() string {
	return NAME
}

// Recover recover old services when master restart
func (e *dockerEngine) Recover() {
	// clean old services in docker mode
	e.clean()
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

// Clean recover all old instances
func (e *dockerEngine) clean() {
	sss := map[string]map[string]attribute{}
	if e.LoadStats(&sss) {
		for sn, instances := range sss {
			for in, instance := range instances {
				id := instance.Container.ID
				if id == "" {
					e.log.Warnf("[%s][%s] container id not found, maybe running mode changed", sn, in)
					continue
				}
				err := e.stopContainer(id)
				if err != nil {
					e.log.Warnf("[%s][%s] failed to stop the old container (%s)", sn, in, id[:12])
				} else {
					e.log.Infof("[%s][%s] old container (%s) stopped", sn, in, id[:12])
				}
				err = e.removeContainer(id)
				if err != nil {
					e.log.Warnf("[%s][%s] failed to remove the old container (%s)", sn, in, id[:12])
				} else {
					e.log.Infof("[%s][%s] old container (%s) removed", sn, in, id[:12])
				}
			}
		}
	}
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
		volumes = append(volumes, fmt.Sprintf(f, v.Path, path.Clean(m.Path)))
	}
	sock := utils.GetEnv(openedge.EnvMasterHostSocket)
	if sock != "" {
		volumes = append(volumes, fmt.Sprintf(fmtVolumeRO, sock, openedge.DefaultSockFile))
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
		Image:        strings.TrimSpace(cfg.Image),
		Env:          utils.AppendEnv(cfg.Env, false),
		Cmd:          cfg.Args,
		ExposedPorts: exposedPorts,
		Labels:       map[string]string{"openedge": "openedge", "service": cfg.Name},
	}
	params.hostConfig = container.HostConfig{
		Binds:        volumes,
		Runtime:      cfg.Runtime,
		PortBindings: portBindings,
		// container is supervised by openedge,
		RestartPolicy: container.RestartPolicy{Name: "no"},
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
	for _, device := range devices {
		deviceParts := strings.Split(device, ":")
		deviceMapping := container.DeviceMapping{}
		switch len(deviceParts) {
		case 1:
			deviceMapping.PathOnHost = deviceParts[0]
			deviceMapping.PathInContainer = deviceParts[0]
			deviceMapping.CgroupPermissions = "mrw"
		case 2:
			deviceMapping.PathOnHost = deviceParts[0]
			deviceMapping.PathInContainer = deviceParts[1]
			deviceMapping.CgroupPermissions = "mrw"
		case 3:
			deviceMapping.PathOnHost = deviceParts[0]
			deviceMapping.PathInContainer = deviceParts[1]
			deviceMapping.CgroupPermissions = deviceParts[2]
		default:
			err = fmt.Errorf("invaild device mapping(%s)", device)
			return
		}
		deviceBindings = append(deviceBindings, deviceMapping)
	}
	return
}

func (e *dockerEngine) Close() error {
	return e.cli.Close()
}
