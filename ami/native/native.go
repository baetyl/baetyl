package native

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	gctx "github.com/baetyl/baetyl-go/v2/context"
	v2context "github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	gHTTP "github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/native"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/kardianos/service"
	"github.com/shirou/gopsutil/v3/cpu"
	gdisk "github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	gnet "github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"

	"github.com/baetyl/baetyl/v2/ami"
	"github.com/baetyl/baetyl/v2/config"
	"github.com/baetyl/baetyl/v2/program"
)

var (
	ErrCreateService = errors.New("failed to create service")
)

const (
	Rows         = 80
	Cols         = 160
	TtySpeed     = 14400
	Term         = "xterm"
	Network      = "tcp"
	LocalAddress = "0.0.0.0"
	PrefixHTTP   = "http://"
	PrefixHTTPS  = "https://"
)

const CmdKillErr = "signal: killed"

func init() {
	ami.Register("native", newNativeImpl)
}

type nativeImpl struct {
	logHostPath   string
	runHostPath   string
	hostPathLib   string
	mapping       *native.ServiceMapping
	portAllocator *native.PortAllocator
	log           *log.Logger
}

func newNativeImpl(cfg config.AmiConfig) (ami.AMI, error) {
	hostPathLib, err := v2context.HostPathLib()
	if err != nil {
		return nil, errors.Trace(err)
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
		hostPathLib:   hostPathLib,
		mapping:       mapping,
		portAllocator: portAllocator,
		log:           log.With(log.Any("ami", "native")),
	}, nil
}

// TODO: impl native UpdateNodeLabels
func (impl *nativeImpl) UpdateNodeLabels(string, map[string]string) error {
	return errors.New("failed to update node label, function has not been implemented")
}

func (impl *nativeImpl) RemoteWebsocket(ctx context.Context, option *ami.DebugOptions, pipe ami.Pipe) error {
	return ami.RemoteWebsocket(ctx, option, pipe)
}

// RemoteCommand Implement of native
func (impl *nativeImpl) RemoteCommand(option *ami.DebugOptions, pipe ami.Pipe) error {
	cfg := &ssh.ClientConfig{
		User: option.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(option.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	server := fmt.Sprintf("%s:%s", option.IP, option.Port)
	conn, err := ssh.Dial(Network, server, cfg)
	if err != nil {
		return errors.Trace(err)
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		return errors.Trace(err)
	}
	defer session.Close()

	session.Stdout = pipe.OutWriter
	session.Stderr = pipe.OutWriter
	session.Stdin = pipe.InReader

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,        // enable echo
		ssh.TTY_OP_ISPEED: TtySpeed, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: TtySpeed, // output speed = 14.4kbaud
	}

	// TODO: support window resize
	if err = session.RequestPty(Term, Rows, Cols, modes); err != nil {
		return errors.Trace(err)
	}
	// Start remote shell
	if err = session.Shell(); err != nil {
		return errors.Trace(err)
	}
	err = session.Wait()
	if err != nil {
		impl.log.Warn("ssh session log out with exception")
	}
	return nil
}

// RemoteLogs use command tail -f
func (impl *nativeImpl) RemoteLogs(option *ami.LogsOptions, pipe ami.Pipe) error {
	logPath := impl.logHostPath
	pathArr := strings.Split(option.Name, ".")
	if len(pathArr) != 5 {
		return errors.Trace(errors.New("log path error"))
	}
	logPath = logPath + "/" + pathArr[0] + "/" + pathArr[1] + "/" + pathArr[2] + "/" + pathArr[3] + "-" + pathArr[4] + ".log"

	tailLines := int64(200)
	if option.TailLines != nil {
		tailLines = *option.TailLines
	}
	cmd := exec.CommandContext(pipe.Ctx, "tail", "-n", strconv.FormatInt(tailLines, 10), "-f", logPath)
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return errors.Trace(err)
	}
	defer stdoutPipe.Close()
	err = cmd.Start()
	if err != nil {
		return errors.Trace(err)
	}
	_, err = io.Copy(pipe.OutWriter, stdoutPipe)
	if err != nil {
		return errors.Trace(err)
	}
	err = cmd.Wait()
	if err != nil {
		if err.Error() == CmdKillErr {
			return nil
		}
		return errors.Trace(err)
	}
	return nil
}

// TODO: impl native RemoteDescribePod
func (impl *nativeImpl) RemoteDescribe(_, _, _ string) (string, error) {
	return "", errors.New("failed to start remote describe pod, function has not been implemented")
}

func (impl *nativeImpl) GetModeInfo() (interface{}, error) {
	return "", nil
}

func (impl *nativeImpl) GetMasterNodeName() string {
	ho, err := host.Info()
	if err != nil {
		return ""
	}
	return ho.Hostname
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
						err = os.WriteFile(filepath.Join(dir, name), []byte(data), 0755)
						if err != nil {
							return errors.Trace(err)
						}
						if name == program.ProgramEntryYaml && prgExec == "" {
							var entry program.Entry
							err = utils.LoadYAML(filepath.Join(dir, name), &entry)
							if err != nil {
								return errors.Trace(err)
							}
							if filepath.IsAbs(entry.Entry) {
								prgExec = filepath.Clean(entry.Entry)
							} else {
								prgExec = filepath.Join(dir, filepath.Join("/", entry.Entry))
							}
						}
					}
				} else if av.Secret != nil {
					vs := secrets[av.Secret.Name]
					for name, data := range vs.Data {
						err = os.WriteFile(filepath.Join(dir, name), data, 0755)
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

			s.Env = setEnv(s.Env, v2context.KeyServiceDynamicPort, strconv.Itoa(port))

			ports = append(ports, port)

			// apply service
			env := []string{
				// MacOS won't set PATH, but function runtimes need it
				fmt.Sprintf("%s=%s", "PATH", os.Getenv("PATH")),
				fmt.Sprintf("%s=%s", v2context.KeyBaetylHostPathLib, impl.hostPathLib),
			}
			for _, item := range s.Env {
				env = append(env, fmt.Sprintf("%s=%s", item.Name, item.Value))
			}

			sysCfg := v2context.SystemConfig{}
			confFilePath := filepath.Join(insDir, program.ProgramConfYaml)
			if utils.FileExists(confFilePath) {
				err = utils.LoadYAML(confFilePath, &sysCfg)
				if err != nil {
					return errors.Trace(err)
				}
			} else {
				err = utils.UnmarshalYAML(nil, &sysCfg)
				if err != nil {
					return errors.Trace(err)
				}
			}

			prgCfg := program.Config{
				Name:        genServiceInstanceName(ns, app.Name, app.Version, s.Name, strconv.Itoa(i)),
				DisplayName: fmt.Sprintf("%s %s", app.Name, s.Name),
				Description: app.Description,
				Dir:         insDir,
				Exec:        prgExec,
				Args:        s.Args,
				Env:         env,
			}
			prgCfg.Logger = sysCfg.Logger
			prgCfg.Logger.Filename = filepath.Join(impl.logHostPath, ns, app.Name, app.Version, fmt.Sprintf("%s-%d.log", s.Name, i))

			prgYml, err := yaml.Marshal(prgCfg)
			if err != nil {
				return errors.Trace(err)
			}
			err = os.WriteFile(filepath.Join(insDir, program.ProgramServiceYaml), prgYml, 0755)
			if err != nil {
				return errors.Trace(err)
			}
			svc, err := service.New(nil, &service.Config{
				Name:             prgCfg.Name,
				DisplayName:      prgCfg.Name,
				Description:      prgCfg.Description,
				WorkingDirectory: insDir,
				Arguments:        []string{"program", insDir},
			})
			if err != nil {
				return errors.Trace(err)
			}
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
			if app.Type == v1.AppTypeContainer {
				err := impl.mapping.SetServicePorts(s.Name, ports)
				if err != nil {
					return errors.Trace(err)
				}
				impl.log.Debug("set applied service ports in mapping files", log.Any("applied service", s.Name), log.Any("ports", ports))
			} else {
				// native function use app name
				err := impl.mapping.SetServicePorts(app.Name, ports)
				if err != nil {
					return errors.Trace(err)
				}
				impl.log.Debug("set applied service ports in mapping files", log.Any("applied service", app.Name), log.Any("ports", ports))
			}
		}
	}
	impl.log.Info("apply an app", log.Any("app", app))
	return nil
}

func (impl *nativeImpl) DeleteApp(ns string, appName string) error {
	// scan app version
	curAppDir := filepath.Join(impl.runHostPath, ns, appName)
	appVerFiles, err := os.ReadDir(curAppDir)
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
		svcFiles, err := os.ReadDir(curAppVerDir)
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
			svcInsFiles, err := os.ReadDir(curSvcDir)
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
				if err != nil {
					return errors.Trace(err)
				}
				var childs *[]ami.ProcessInfo
				if runtime.GOOS == "windows" {
					pid, perr := svc.GetPid()
					if perr != nil {
						impl.log.Warn("failed to get svc pid", log.Error(err))
					}
					childs, perr = getChildProcessInfo(pid)
					if perr != nil {
						impl.log.Warn("failed to get child pid", log.Error(err))
					}
				}
				if err = svc.Stop(); err != nil {
					impl.log.Warn("failed to stop old app", log.Error(err))
				}
				if err = svc.Uninstall(); err != nil {
					impl.log.Warn("failed to uninstall old app", log.Error(err))
				}

				if runtime.GOOS == "windows" {
					for _, p := range *childs {
						proc, err := process.NewProcess(p.Pid)
						if err == nil {
							proc.Kill()
							impl.log.Warn("svc child process killed", log.Any("name", p.Name), log.Any("pid", p.Pid))
							time.Sleep(100 * time.Millisecond)
						}
					}
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
			impl.log.Debug("delete applied service ports in mapping files", log.Any("applied service", curSvcName))
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
	appFiles, err := os.ReadDir(curNsPath)
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
		appVerFiles, err := os.ReadDir(curAppPath)
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
			svcFiles, err := os.ReadDir(curAppVerPath)
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
				svcInsFiles, err := os.ReadDir(curSvcPath)
				if err != nil {
					return nil, errors.Trace(err)
				}
				for _, svcInsFile := range svcInsFiles {
					if !svcInsFile.IsDir() {
						continue
					}
					var pid uint32
					curSvcIns := svcInsFile.Name()
					curPrgName := genServiceInstanceName(ns, curAppName, curAppVer, curSvcName, curSvcIns)
					curInsStats := map[string]v1.InstanceStats{}
					mainInsStats := v1.InstanceStats{
						ServiceName: curSvcName,
						Name:        curPrgName,
					}
					svc, err := service.New(nil, &service.Config{
						Name:             curPrgName,
						WorkingDirectory: svcInsFile.Name(),
					})
					if err != nil {
						mainInsStats.Status = v1.Unknown
						mainInsStats.Cause = err.Error()
					}
					if svc != nil {
						status, err := svc.Status()
						if err != nil {
							mainInsStats.Status = v1.Unknown
							mainInsStats.Cause += err.Error()
						} else {
							mainInsStats.Status = prgStatusToSpecStatus(status)
						}
						pid, err = svc.GetPid()
						if err != nil {
							mainInsStats.Status = v1.Unknown
							mainInsStats.Cause += err.Error()
						}
						mainInsStats.Pid = int32(pid)
						ppid, err := getPPID(pid)
						if err != nil {
							mainInsStats.Status = v1.Unknown
							mainInsStats.Cause += err.Error()
						}
						mainInsStats.PPid = ppid
						usage, err := getServiceInsStats(pid)
						if err != nil {
							mainInsStats.Status = v1.Unknown
							mainInsStats.Cause += err.Error()
						} else {
							mainInsStats.Usage = usage
						}
					} else {
						mainInsStats.Status = v1.Unknown
						mainInsStats.Cause += ErrCreateService.Error()
					}
					curInsStats[curPrgName] = mainInsStats

					getChildInsStats(curInsStats, pid, curPrgName)

					curAppStats.InstanceStats = curInsStats
				}
			}
			curAppStats.Status = getAppStatus(curAppStats.InstanceStats)

			if len(curAppStats.InstanceStats) > 0 {
				stats = append(stats, curAppStats)
			}
		}
	}
	return stats, nil
}

func getChildInsStats(curInsStats map[string]v1.InstanceStats, pid uint32, curPrgName string) {
	childs, err := getChildProcessInfo(pid)
	if err != nil {
		return
	}

	mainInsStats, ok := curInsStats[curPrgName]
	if !ok {
		return
	}

	for _, cc := range *childs {
		usage, err := getServiceInsStats(uint32(cc.Pid))
		if err != nil {
			return
		}

		curInsStats[cc.Name] = v1.InstanceStats{
			Name:        cc.Name,
			ServiceName: mainInsStats.ServiceName,
			Usage:       usage,
			Status:      mainInsStats.Status,
			Pid:         cc.Pid,
			PPid:        cc.Ppid,
		}
	}
}

func getChildProcessInfo(pid uint32) (*[]ami.ProcessInfo, error) {
	var pProc *process.Process
	var childs []ami.ProcessInfo
	processes, err := process.Processes()
	if err != nil {
		return nil, err
	}
	for _, p := range processes {
		if p.Pid == int32(pid) {
			pProc = p
			break
		}
	}
	if pProc == nil {
		return nil, errors.Errorf("failed to get process of pid %d", pid)
	}

	getChild(pProc, &childs)

	return &childs, nil
}

func getPPID(pid uint32) (int32, error) {
	processes, err := process.Processes()
	if err != nil {
		return 1, err
	}
	for _, p := range processes {
		if p.Pid == int32(pid) {
			return p.Ppid()
		}
	}

	return 1, nil
}

func getChild(p *process.Process, childs *[]ami.ProcessInfo) {
	c, _ := p.Children()
	if len(c) == 0 {
		return
	}
	for _, cc := range c {
		var procInfo ami.ProcessInfo
		procInfo.Pid = cc.Pid
		name, err := cc.Name()
		if err != nil {
			return
		}
		cPPid, err := cc.Ppid()
		if err != nil {
			return
		}
		procInfo.Name = name
		procInfo.Ppid = cPPid

		*childs = append(*childs, procInfo)
		getChild(cc, childs)
	}
}

func getServiceInsStats(pid uint32) (map[string]string, error) {
	usage := map[string]string{}
	proc, err := process.NewProcess(int32(pid))
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

func getAppStatus(infos map[string]v1.InstanceStats) v1.Status {
	var pending = false
	for _, info := range infos {
		if info.Status == v1.Pending {
			pending = true
		} else if info.Status == v1.Unknown {
			return info.Status
		}
	}
	if pending {
		return v1.Pending
	}
	return v1.Running
}

func (impl *nativeImpl) CollectNodeInfo() (map[string]interface{}, error) {
	ho, err := host.Info()
	if err != nil {
		return nil, errors.Trace(err)
	}
	plat := v2context.Platform()
	ias, err := net.InterfaceAddrs()
	if err != nil {
		return nil, errors.Trace(err)
	}
	var addrs []string
	for _, ia := range ias {
		if ipnet, ok := ia.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				addrs = append(addrs, ipnet.IP.String())
			}
		}
	}
	return map[string]interface{}{
		ho.Hostname: &v1.NodeInfo{
			Arch:       runtime.GOARCH,
			OS:         runtime.GOOS,
			Variant:    plat.Variant,
			SystemUUID: ho.HostID,
			Hostname:   ho.Hostname,
			Role:       "master",
			Address:    strings.Join(addrs, ","),
		},
	}, nil
}

func (impl *nativeImpl) CollectNodeStats() (map[string]interface{}, error) {
	stats := &v1.NodeStats{
		Usage:    map[string]string{},
		Capacity: map[string]string{},
		Percent:  map[string]string{},
		NetIO:    map[string]string{},
	}
	ho, err := host.Info()
	if err != nil {
		return nil, errors.Trace(err)
	}
	infos, err := cpu.Info()
	if err != nil {
		return nil, errors.Trace(err)
	}
	var cores int32
	for _, info := range infos {
		cores += info.Cores
	}
	percent, err := cpu.Percent(0, false)
	stats.Capacity["cpu"] = strconv.FormatInt(int64(cores), 10)
	if len(percent) >= 1 {
		usage := float64(cores) * percent[0] / 100
		stats.Usage["cpu"] = strconv.FormatFloat(usage, 'f', 3, 64)
	}
	// TODO replace with more appropriate stats
	me, err := mem.VirtualMemory()
	if err != nil {
		return nil, errors.Trace(err)
	}
	stats.Capacity["memory"] = strconv.FormatUint(me.Total, 10)
	stats.Usage["memory"] = strconv.FormatUint(me.Used, 10)

	disk, err := gdisk.Usage("/")
	if err != nil {
		return nil, errors.Trace(err)
	}
	stats.Capacity["disk"] = strconv.FormatUint(disk.Total, 10)
	stats.Usage["disk"] = strconv.FormatUint(disk.Used, 10)
	diskPercent := float64(disk.Used) / float64(disk.Total)
	stats.Percent["disk"] = strconv.FormatFloat(diskPercent, 'f', -1, 64)

	diskPartitions, err := gdisk.Partitions(false)
	if err != nil {
		return nil, errors.Trace(err)
	}
	for _, p := range diskPartitions {
		pUsage, pErr := gdisk.Usage(p.Mountpoint)
		if pErr != nil {
			return nil, errors.Trace(pErr)
		}
		stats.Capacity["disk_"+p.Mountpoint] = strconv.FormatUint(pUsage.Total, 10)
		stats.Usage["disk_"+p.Mountpoint] = strconv.FormatUint(pUsage.Used, 10)
		pDiskPercent := float64(pUsage.Used) / float64(pUsage.Total)
		stats.Percent["disk_"+p.Mountpoint] = strconv.FormatFloat(pDiskPercent, 'f', -1, 64)
	}

	netIO, err := gnet.IOCounters(false)
	if err != nil {
		return nil, errors.Trace(err)
	}
	// sleep 1s to get net speed
	time.Sleep(1000 * time.Millisecond)
	netIOSecond, err := gnet.IOCounters(false)
	if err != nil {
		return nil, errors.Trace(err)
	}
	InBytes := netIOSecond[0].BytesRecv - netIO[0].BytesRecv
	OutBytes := netIOSecond[0].BytesSent - netIO[0].BytesSent
	InPackets := netIOSecond[0].PacketsRecv - netIO[0].PacketsRecv
	OutPackets := netIOSecond[0].PacketsSent - netIO[0].PacketsSent

	stats.NetIO["netBytesSent"] = strconv.FormatUint(OutBytes, 10)
	stats.NetIO["netBytesRecv"] = strconv.FormatUint(InBytes, 10)
	stats.NetIO["netPacketsSent"] = strconv.FormatUint(OutPackets, 10)
	stats.NetIO["netPacketsRecv"] = strconv.FormatUint(InPackets, 10)

	var gpuExts map[string]interface{}
	if extension, ok := ami.Hooks[ami.BaetylGPUStatsExtension]; ok {
		collectStatsExt, ok := extension.(ami.CollectStatsExtFunc)
		if ok {
			gpuExts, err = collectStatsExt(gctx.RunModeNative)
			if err != nil {
				impl.log.Warn("failed to collect gpu stats", log.Error(errors.Trace(err)))
			}
			impl.log.Debug("collect gpu stats successfully", log.Any("gpuStats", gpuExts))
		} else {
			impl.log.Warn("invalid collecting gpu stats function")
		}
	}
	var nodeStatsMerge map[string]interface{}
	if len(gpuExts) > 0 {
		if ext, ok := gpuExts[ho.Hostname]; ok && ext != nil {
			nodeStatsMerge = ext.(map[string]interface{})
		}
	}
	stats.Extension = nodeStatsMerge
	// TODO add pressure flags
	return map[string]interface{}{ho.Hostname: stats}, nil
}

func (impl *nativeImpl) FetchLog(namespace, service string, tailLines, sinceSeconds int64) (io.ReadCloser, error) {
	panic("implement me")
}

func (impl *nativeImpl) RPCApp(url string, req *v1.RPCRequest) (*v1.RPCResponse, error) {
	if strings.HasPrefix(url, PrefixHTTPS) {
		url = PrefixHTTPS + LocalAddress + req.Params
	} else {
		url = PrefixHTTP + LocalAddress + req.Params
	}
	ops := gHTTP.NewClientOptions()
	ops.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	cli := gHTTP.NewClient(ops)
	impl.log.Debug("rpc http start", log.Any("url", url), log.Any("method", req.Method))

	var buf []byte
	if req.Body != nil {
		buf = []byte(fmt.Sprintf("%v", req.Body))
	}
	res, err := cli.SendUrl(strings.ToUpper(req.Method), url, bytes.NewReader(buf), req.Header)
	if err != nil {
		return nil, errors.Trace(err)
	}

	response := &v1.RPCResponse{
		StatusCode: res.StatusCode,
		Header:     res.Header,
	}
	response.Body, err = gHTTP.HandleResponse(res)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return response, nil
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
