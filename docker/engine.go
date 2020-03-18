package docker

import (
	"context"
	"fmt"
	"path"
	"runtime"
	"strings"
	"sync"

	"github.com/baetyl/baetyl"
	"github.com/baetyl/baetyl/logger"
	schema "github.com/baetyl/baetyl/schema/v3"
	"github.com/baetyl/baetyl/utils"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/sysinfo"
	"github.com/docker/go-connections/nat"
	cmap "github.com/orcaman/concurrent-map"
)

const engineName = "docker"

func init() {
	baetyl.RegisterEngine(engineName, newEngine)
}

func newEngine(ctx baetyl.Context) (baetyl.Engine, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithVersion(ctx.Config().Docker.APIVersion))
	if err != nil {
		return nil, err
	}
	info, err := cli.Info(context.Background())
	if err != nil {
		return nil, err
	}
	e := &engine{
		ctx:      ctx,
		svcs:     cmap.New(),
		cli:      cli,
		info:     info,
		networks: make(map[string]string),
		log:      ctx.Logger().WithField("engine", engineName),
	}
	return e, nil
}

type engine struct {
	ctx      baetyl.Context
	svcs     cmap.ConcurrentMap
	cli      *client.Client
	info     types.Info
	networks map[string]string
	tomb     utils.Tomb
	log      logger.Logger
}

func (e *engine) Name() string {
	return engineName
}

func (e *engine) OSType() string {
	return e.info.OSType
}

func (e *engine) Close() error {
	e.clean()
	return e.cli.Close()
}

func (e *engine) clean() {
	var wg sync.WaitGroup
	for item := range e.svcs.Iter() {
		wg.Add(1)
		go func(e *engine, svc *service) {
			svc.stop()
			wg.Done()
		}(e, item.Val.(*service))
	}
	wg.Wait()
	e.log.Infoln("all services stopped")
}

func (e *engine) Apply(ctx context.Context, appcfg *schema.ComposeAppConfig) error {
	defer e.clean()
	defer e.log.Infoln("complete apply appconfig")
	e.log.Infoln("begin apply appconfig")
	e.prepare(*appcfg)
	e.log.Infoln("prepare done")
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	for _, name := range sortedServices(appcfg.Services) {
		e.log.Infof("operating service (%s)", name)
		s := appcfg.Services[name]
		if s.ContainerName != "" {
			name = s.ContainerName
		}
		svc, err := e.createService(name, &s, appcfg.Volumes)
		if err != nil {
			e.log.Warnf("failed to create service (%s)", name)
			return err
		}
		e.svcs.Set(name, svc)
		go svc.run()
	}
	<-ctx.Done()
	return ctx.Err()
}

func (e *engine) UpdateStats() {
	// TODO
}

func sortedServices(services map[string]schema.ComposeService) []string {
	g := map[string][]string{}
	inDegrees := map[string]int{}
	res := []string{}
	for name, s := range services {
		for _, r := range s.DependsOn {
			if g[r] == nil {
				g[r] = []string{}
			}
			g[r] = append(g[r], name)
		}
		inDegrees[name] = len(s.DependsOn)
	}
	queue := []string{}
	for n, i := range inDegrees {
		if i == 0 {
			queue = append(queue, n)
			inDegrees[n] = -1
		}
	}
	for len(queue) > 0 {
		i := queue[0]
		res = append(res, i)
		queue = queue[1:]
		for _, v := range g[i] {
			inDegrees[v]--
			if inDegrees[v] == 0 {
				inDegrees[v] = -1
				queue = append(queue, v)
			}
		}
	}
	return res
}

func (e *engine) prepare(cfg schema.ComposeAppConfig) {
	var wg sync.WaitGroup
	ss := cfg.Services
	for _, s := range ss {
		wg.Add(1)
		go func(i string, w *sync.WaitGroup) {
			defer w.Done()
			e.pullImage(i)
		}(s.Image, &wg)
	}
	wg.Add(1)
	go func(nw map[string]schema.ComposeNetwork, w *sync.WaitGroup) {
		defer w.Done()
		e.initNetworks(nw)
	}(cfg.Networks, &wg)

	wg.Add(1)
	go func(vs map[string]schema.ComposeVolume, w *sync.WaitGroup) {
		defer w.Done()
		e.initVolumes(vs)
	}(cfg.Volumes, &wg)
	wg.Wait()
}

func (e *engine) createService(name string, meta *schema.ComposeService, vs map[string]schema.ComposeVolume) (*service, error) {
	if runtime.GOOS == "linux" && meta.Resources.CPU.Cpus > 0 {
		sysInfo := sysinfo.New(true)
		if !sysInfo.CPUCfsPeriod || !sysInfo.CPUCfsQuota {
			e.log.Warnf("configuration 'resources.cpu.cpus' of service (%s) is ignored, because host kernel does not support CPU cfs period/quota or the cgroup is not mounted.", name)
			meta.Resources.CPU.Cpus = 0
		}
	}
	binds := make([]string, 0)
	volumes := map[string]struct{}{}
	for _, m := range meta.Volumes {
		if _, ok := vs[m.Source]; !ok {
			if m.Type == "volume" {
				return nil, fmt.Errorf("volume '%s' not found", m.Source)
			}
		}
		f := fmtVolumeRW
		if m.ReadOnly {
			f = fmtVolumeRO
		}
		binds = append(binds, fmt.Sprintf(f, m.Source, path.Clean(m.Target)))
		volumes[m.Target] = struct{}{}
	}
	exposedPorts, portBindings, err := nat.ParsePortSpecs(meta.Ports)
	if err != nil {
		return nil, err
	}
	deviceBindings, err := e.parseDeviceSpecs(meta.Devices)
	if err != nil {
		return nil, err
	}
	var template template
	template.config = container.Config{
		Image:        strings.TrimSpace(meta.Image),
		Env:          utils.AppendEnv(meta.Environment.Envs, false),
		Cmd:          meta.Command.Cmd,
		Hostname:     meta.Hostname,
		ExposedPorts: exposedPorts,
		Volumes:      volumes,
		Labels:       map[string]string{"baetyl": "baetyl", "service": name},
	}
	endpointsConfig := map[string]*network.EndpointSettings{}
	if meta.NetworkMode != "" {
		if len(meta.Networks.ServiceNetworks) > 0 {
			return nil, fmt.Errorf("'network_mode' and 'networks' cannot be combined")
		}
	} else {
		for networkName, networkInfo := range meta.Networks.ServiceNetworks {
			meta.NetworkMode = networkName
			endpointsConfig[networkName] = &network.EndpointSettings{
				NetworkID: e.networks[networkName],
				Aliases:   networkInfo.Aliases,
				IPAddress: networkInfo.Ipv4Address,
			}
		}
		if meta.NetworkMode == "" {
			meta.NetworkMode = defaultNetworkName
		}
	}
	template.networkConfig = network.NetworkingConfig{
		EndpointsConfig: endpointsConfig,
	}
	template.hostConfig = container.HostConfig{
		Binds:        binds,
		Runtime:      meta.Runtime,
		PortBindings: portBindings,
		NetworkMode:  container.NetworkMode(meta.NetworkMode),
		// container is supervised by baetyl,
		RestartPolicy: container.RestartPolicy{Name: "no"},
		Resources: container.Resources{
			CpusetCpus: meta.Resources.CPU.SetCPUs,
			NanoCPUs:   int64(meta.Resources.CPU.Cpus * 1e9),
			Memory:     meta.Resources.Memory.Limit,
			MemorySwap: meta.Resources.Memory.Swap,
			PidsLimit:  &meta.Resources.Pids.Limit,
			Devices:    deviceBindings,
		},
	}
	s := &service{
		name:     name,
		log:      e.log.WithField("service", name),
		e:        e,
		meta:     meta,
		template: template,
		pods:     cmap.New(),
		sig:      make(chan struct{}, 1),
	}
	return s, nil
}

func (e *engine) parseDeviceSpecs(devices []string) (deviceBindings []container.DeviceMapping, err error) {
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
