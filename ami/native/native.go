package native

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/native"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/kardianos/service"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
	"gopkg.in/yaml.v2"

	"github.com/baetyl/baetyl/v2/ami"
	"github.com/baetyl/baetyl/v2/config"
	"github.com/baetyl/baetyl/v2/engine"
	"github.com/baetyl/baetyl/v2/program"
)

func init() {
	ami.Register("native", newNativeImpl)
}

type nativeImpl struct {
	logHostPath    string
	runHostPath    string
	mapping        *native.ServiceMapping
	portAllocator  *native.PortAllocator
	hostPathLibRun string
	log            *log.Logger
}

func newNativeImpl(cfg config.AmiConfig) (ami.AMI, error) {
	var hostPathLib string
	if val := os.Getenv(context.KeyBaetylHostPathLib); val == "" {
		err := os.Setenv(context.KeyBaetylHostPathLib, engine.DefaultHostPathLib)
		if err != nil {
			return nil, errors.Trace(err)
		}
		hostPathLib = engine.DefaultHostPathLib
	} else {
		hostPathLib = val
	}
	portAllocator, err := native.NewPortAllocator(cfg.Native.PortsRange.Start, cfg.Native.PortsRange.End)
	if err != nil {
		return nil, errors.Trace(err)
	}
	mapping, err := native.NewServiceMapping()
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &nativeImpl{
		logHostPath:   filepath.Join(hostPathLib, "log"),
		runHostPath:   filepath.Join(hostPathLib, "run"),
		mapping:       mapping,
		portAllocator: portAllocator,
		log:           log.With(log.Any("ami", "native")),
	}, nil
}

// TODO: impl native RemoteCommand
func (impl *nativeImpl) RemoteCommand(option ami.DebugOptions, stdin io.Reader, stdout, stderr io.Writer) error {
	panic("impl me")
}

func (impl *nativeImpl) ApplyApp(ns string, app v1.Application, configs map[string]v1.Configuration, secrets map[string]v1.Secret) error {
	err := impl.DeleteApp(ns, app.Name)
	if err != nil {
		impl.log.Warn("failed to delete old app", log.Error(err))
	}

	appDir := filepath.Join(impl.runHostPath, ns, app.Name, app.Version)
	err = os.MkdirAll(appDir, 0755)
	if err != nil {
		return errors.Trace(err)
	}
	avs := map[string]v1.Volume{}
	for _, v := range app.Volumes {
		avs[v.Name] = v
	}

	for _, s := range app.Services {
		var ports []int
		for i := 1; i <= s.Replica; i++ {
			var prgExec string

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
					mp := filepath.Join(insDir, filepath.Join("/", vm.MountPath))
					if err = os.MkdirAll(filepath.Dir(mp), 0755); err != nil {
						return errors.Trace(err)
					}
					if err = os.Symlink(av.HostPath.Path, mp); err != nil {
						return errors.Trace(err)
					}

					impl.log.Info("mount a volume", log.Any("vm", vm))
					if vm.MountPath == program.ProgramBinPath {
						var entry program.Entry
						err = utils.LoadYAML(filepath.Join(mp, program.ProgramEntryYaml), &entry)
						if err != nil {
							return errors.Trace(err)
						}
						if filepath.IsAbs(entry.Entry) {
							prgExec = filepath.Clean(entry.Entry)
						} else {
							prgExec = filepath.Join(mp, filepath.Join("/", entry.Entry))
						}
					}
					continue
				}

				// create mount path
				dir := filepath.Join(insDir, vm.MountPath)
				if err = os.MkdirAll(dir, 0755); err != nil {
					return errors.Trace(err)
				}

				if av.Config != nil {
					vc := configs[av.Config.Name]
					for name, data := range vc.Data {
						err = ioutil.WriteFile(filepath.Join(dir, name), []byte(data), 0755)
						if err != nil {
							return errors.Trace(err)
						}
					}
				} else if av.Secret != nil {
					vs := secrets[av.Secret.Name]
					for name, data := range vs.Data {
						err = ioutil.WriteFile(filepath.Join(dir, name), data, 0755)
						if err != nil {
							return errors.Trace(err)
						}
					}
				}
			}

			if prgExec == "" {
				err = impl.DeleteApp(ns, app.Name)
				if err != nil {
					impl.log.Warn("failed to delete new app", log.Error(err))
				}
				return errors.Errorf("no program executable, the program config may not be mounted")
			}

			port, err := impl.portAllocator.Allocate()
			if err != nil {
				return errors.Trace(err)
			}

			s.Env = setEnv(s.Env, context.KeyServiceDynamicPort, strconv.Itoa(port))

			ports = append(ports, port)

			// apply service
			var env []string
			for _, item := range s.Env {
				env = append(env, fmt.Sprintf("%s=%s", item.Name, item.Value))
			}
			prgCfg := program.Config{
				Name:        genServiceInstanceName(ns, app.Name, app.Version, s.Name, strconv.Itoa(i)),
				DisplayName: fmt.Sprintf("%s %s", app.Name, s.Name),
				Description: app.Description,
				Dir:         insDir,
				Exec:        prgExec,
				Args:        s.Args,
				Env:         env,
				Logger: log.Config{
					Level:    "debug",
					Filename: filepath.Join(impl.logHostPath, ns, app.Name, app.Version, fmt.Sprintf("%s-%d.log", s.Name, i)),
				},
			}
			prgYml, err := yaml.Marshal(prgCfg)
			if err != nil {
				return errors.Trace(err)
			}
			err = ioutil.WriteFile(filepath.Join(insDir, program.ProgramServiceYaml), prgYml, 0755)
			if err != nil {
				return errors.Trace(err)
			}
			svc, err := service.New(nil, &service.Config{
				Name:             prgCfg.Name,
				Description:      prgCfg.Description,
				WorkingDirectory: insDir,
				Arguments:        []string{"program"},
			})
			err = svc.Install()
			if err != nil {
				svc.Uninstall()
				err = svc.Install()
				if err != nil {
					return errors.Trace(err)
				}
			}
			err = svc.Start()
			if err != nil {
				svc.Stop()
				err = svc.Start()
				if err != nil {
					return errors.Trace(err)
				}
			}
		}

		if len(ports) > 0 {
			err := impl.mapping.SetServicePorts(s.Name, ports)
			if err != nil {
				return errors.Trace(err)
			}
		}
	}
	impl.log.Info("apply an app", log.Any("app", app))
	return nil
}

func (impl *nativeImpl) DeleteApp(ns string, appName string) error {
	// scan app version
	curAppDir := filepath.Join(impl.runHostPath, ns, appName)
	appVerFiles, err := ioutil.ReadDir(curAppDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return errors.Trace(err)
	}
	for _, appVerFile := range appVerFiles {
		if !appVerFile.IsDir() {
			continue
		}
		// scan service
		curAppVer := appVerFile.Name()
		curAppVerDir := filepath.Join(curAppDir, curAppVer)
		svcFiles, err := ioutil.ReadDir(curAppVerDir)
		if err != nil {
			return errors.Trace(err)
		}
		for _, svcFile := range svcFiles {
			if !svcFile.IsDir() {
				continue
			}
			// scan service instance
			curSvcName := svcFile.Name()
			curSvcDir := filepath.Join(curAppVerDir, curSvcName)
			svcInsFiles, err := ioutil.ReadDir(curSvcDir)
			if err != nil {
				return errors.Trace(err)
			}
			for _, svcInsFile := range svcInsFiles {
				if !svcInsFile.IsDir() {
					continue
				}
				curSvcIns := svcInsFile.Name()
				curSvcInsDir := filepath.Join(curSvcDir, curSvcIns)
				svc, err := service.New(nil, &service.Config{
					Name:             genServiceInstanceName(ns, appName, curAppVer, curSvcName, curSvcIns),
					WorkingDirectory: svcInsFile.Name(),
				})
				if err = svc.Uninstall(); err != nil {
					impl.log.Warn("failed to uninstall old app", log.Error(err))
				}
				err = os.RemoveAll(curSvcInsDir)
				if err != nil {
					return errors.Trace(err)
				}
			}
			err = os.RemoveAll(curSvcDir)
			if err != nil {
				return errors.Trace(err)
			}

			err = impl.mapping.DeleteServicePorts(curSvcName)
			if err != nil {
				return errors.Trace(err)
			}
		}
		err = os.RemoveAll(curAppVerDir)
		if err != nil {
			return errors.Trace(err)
		}
	}
	return errors.Trace(os.RemoveAll(curAppDir))
}

func (impl *nativeImpl) StatsApps(ns string) ([]v1.AppStats, error) {
	var stats []v1.AppStats
	if !utils.DirExists(impl.runHostPath) {
		return stats, nil
	}

	curNsPath := filepath.Join(impl.runHostPath, ns)
	if !utils.DirExists(curNsPath) {
		return stats, nil
	}

	// scan app
	appFiles, err := ioutil.ReadDir(curNsPath)
	if err != nil {
		return nil, errors.Trace(err)
	}
	for _, appFile := range appFiles {
		if !appFile.IsDir() {
			continue
		}
		curAppName := appFile.Name()
		curAppPath := filepath.Join(curNsPath, curAppName)
		if !utils.DirExists(curAppPath) {
			continue
		}
		// scan app version
		appVerFiles, err := ioutil.ReadDir(curAppPath)
		if err != nil {
			return nil, errors.Trace(err)
		}
		for _, appVerFile := range appVerFiles {
			if !appVerFile.IsDir() {
				continue
			}

			curAppStats := v1.AppStats{}
			curAppStats.Name = appFile.Name()
			curAppStats.Version = appVerFile.Name()
			curAppStats.InstanceStats = map[string]v1.InstanceStats{}

			curAppVer := appVerFile.Name()
			curAppVerPath := filepath.Join(curAppPath, curAppVer)
			if !utils.DirExists(curAppVerPath) {
				continue
			}
			// scan service
			svcFiles, err := ioutil.ReadDir(curAppVerPath)
			if err != nil {
				return nil, errors.Trace(err)
			}
			for _, svcFile := range svcFiles {
				if !svcFile.IsDir() {
					continue
				}

				curSvcName := svcFile.Name()
				curSvcPath := filepath.Join(curAppVerPath, curSvcName)
				if !utils.DirExists(curSvcPath) {
					continue
				}
				// scan service instance
				svcInsFiles, err := ioutil.ReadDir(curSvcPath)
				if err != nil {
					return nil, errors.Trace(err)
				}
				for _, svcInsFile := range svcInsFiles {
					if !svcInsFile.IsDir() {
						continue
					}

					curSvcIns := svcInsFile.Name()
					curPrgName := genServiceInstanceName(ns, curAppName, curAppVer, curSvcName, curSvcIns)
					svc, err := service.New(nil, &service.Config{
						Name:             curPrgName,
						WorkingDirectory: svcInsFile.Name(),
					})
					curInsStats := v1.InstanceStats{
						ServiceName: curSvcName,
						Name:        curPrgName,
					}
					status, err := svc.Status()
					if err != nil {
						curInsStats.Status = v1.Unknown
						curInsStats.Cause = err.Error()
					} else {
						curInsStats.Status = prgStatusToSpecStatus(status)
					}
					usage, err := getServiceInsStats(svc)
					if err != nil {
						curInsStats.Cause += err.Error()
					} else {
						curInsStats.Usage = usage
					}
					curAppStats.InstanceStats[curPrgName] = curInsStats
				}
			}

			if len(curAppStats.InstanceStats) > 0 {
				stats = append(stats, curAppStats)
			}
		}
	}
	return stats, nil
}

func getServiceInsStats(svc service.Service) (map[string]string, error) {
	usage := map[string]string{}
	pid, err := svc.GetPid()
	if err != nil {
		return nil, errors.Trace(err)
	}
	proc, err := process.NewProcess(pid)
	if err != nil {
		return nil, errors.Trace(err)
	}
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, errors.Trace(err)
	}
	mPercent, err := proc.MemoryPercent()
	if err != nil {
		return nil, errors.Trace(err)
	}
	usedMem := uint64(float32(memInfo.Total) * mPercent / 100)
	usage["memory"] = strconv.FormatUint(usedMem, 10)

	cpuinfos, err := cpu.Info()
	if len(cpuinfos) < 1 {
		return nil, errors.Errorf("failed to get cpu info")
	}
	cPercent, err := proc.CPUPercent()
	if err != nil {
		return nil, errors.Trace(err)
	}
	core := float64(cpuinfos[0].Cores) * cPercent / 100
	usage["cpu"] = strconv.FormatFloat(core, 'f', 3, 64)
	return usage, nil
}

func prgStatusToSpecStatus(status service.Status) v1.Status {
	switch status {
	case service.StatusRunning:
		return v1.Running
	case service.StatusStopped:
		return v1.Pending
	default:
		return v1.Unknown
	}
}

func (impl *nativeImpl) CollectNodeInfo() (*v1.NodeInfo, error) {
	ho, err := host.Info()
	if err != nil {
		return nil, err
	}
	plat := context.Platform()
	// TODO add address
	return &v1.NodeInfo{
		Arch:     runtime.GOARCH,
		OS:       runtime.GOOS,
		Variant:  plat.Variant,
		HostID:   ho.HostID,
		Hostname: ho.Hostname,
	}, nil
}

func (impl *nativeImpl) CollectNodeStats() (*v1.NodeStats, error) {
	stats := &v1.NodeStats{
		Usage:    map[string]string{},
		Capacity: map[string]string{},
	}
	cpuinfos, err := cpu.Info()
	if err != nil {
		return nil, errors.Trace(err)
	}
	cPercent, err := cpu.Percent(0, false)
	if len(cpuinfos) >= 1 {
		cores := int(cpuinfos[0].Cores)
		stats.Capacity["cpu"] = strconv.Itoa(cores)
		if len(cPercent) >= 1 {
			usage := float64(cores) * cPercent[0] / 100
			stats.Usage["cpu"] = strconv.FormatFloat(usage, 'f', 3, 64)
		}
	}

	// TODO replace with more appropriate stats
	me, err := mem.VirtualMemory()
	if err != nil {
		return nil, errors.Trace(err)
	}
	stats.Capacity["memory"] = strconv.FormatUint(me.Total, 10)
	stats.Usage["memory"] = strconv.FormatUint(me.Used, 10)

	// TODO add pressure flags
	return stats, nil
}

func (impl *nativeImpl) FetchLog(namespace, service string, tailLines, sinceSeconds int64) (io.ReadCloser, error) {
	panic("implement me")
}

func genServiceInstanceName(ns, appName, appVersion, svcName, instanceID string) string {
	return fmt.Sprintf("%s.%s.%s.%s.%s", ns, appName, appVersion, svcName, instanceID)
}

func setEnv(env []v1.Environment, key, value string) []v1.Environment {
	for i := 0; i < len(env); i++ {
		if env[i].Name == key {
			env[i].Value = value
			return env
		}
	}

	env = append(env, v1.Environment{
		Name:  key,
		Value: value,
	})
	return env
}
