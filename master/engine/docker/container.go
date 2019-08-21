package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"runtime"
	"time"

	"github.com/baidu/openedge/master/engine"
	"github.com/baidu/openedge/sdk/openedge-go"
	"github.com/baidu/openedge/utils"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/pkg/stdcopy"
)

const defaultNetworkName = "openedge"

type containerConfigs struct {
	config        container.Config
	hostConfig    container.HostConfig
	networkConfig network.NetworkingConfig
}

func (e *dockerEngine) initNetworks(networks map[string]openedge.NetworkInfo) error {
	// add docker default network, including bridge, host, none
	ctx := context.Background()
	args := filters.NewArgs()
	args.Add("type", "builtin")
	nws, err := e.cli.NetworkList(ctx, types.NetworkListOptions{Filters: args})
	if err != nil {
		e.log.WithError(err).Errorf("failed to list builtin networks")
		return err
	}
	for _, builtinNet := range nws {
		e.networks[builtinNet.Name] = builtinNet.ID
	}

	args = filters.NewArgs()
	args.Add("type", "custom")
	nws, err = e.cli.NetworkList(ctx, types.NetworkListOptions{Filters: args})
	if err != nil {
		e.log.WithError(err).Errorf("failed to list custom networks")
		return err
	}
	nwMap := map[string]types.NetworkResource{}
	for _, val := range nws {
		nwMap[val.Name] = val
	}
	// add openedge as default network
	if networks == nil{
		networks = make(map[string]openedge.NetworkInfo)
	}
	networks[defaultNetworkName] = openedge.NetworkInfo {
		Driver: "bridge",
	}
	for networkName, network := range networks {
		if network.Driver == "" {
			network.Driver = "bridge"
		}
		if nw, ok := nwMap[networkName]; ok {
			if nw.Driver != network.Driver {
				return fmt.Errorf("network (%s:%s) with different driver exists", nw.ID[:12], networkName)
			}
			e.networks[networkName] = nw.ID
			e.log.Debugf("network (%s:%s) exists", nw.ID[:12], networkName)
		} else {
			networkParams := types.NetworkCreate {
				Driver: network.Driver,
				Options: network.DriverOptions,
				Scope: "local",
				Labels: network.Labels,
			}
			nw, err := e.cli.NetworkCreate(ctx, networkName, networkParams)
			if err != nil {
				e.log.WithError(err).Errorf("failed to create network (%s)", networkName)
				return err
			}
			if nw.Warning != "" {
				e.log.Warnf(nw.Warning)
			}
			e.networks[networkName] = nw.ID
			e.log.Debugf("network (%s:%s) created", e.networks[networkName][:12], networkName)
		}
	}
	return nil
}

func (e *dockerEngine) pullImage(name string) error {
	out, err := e.cli.ImagePull(context.Background(), name, types.ImagePullOptions{})
	if err != nil {
		e.log.WithError(err).Warnf("failed to pull image (%s)", name)
		return err
	}
	defer out.Close()
	io.Copy(ioutil.Discard, out)
	e.log.Debugf("image (%s) pulled", name)
	return nil
}

func (e *dockerEngine) startContainer(name string, cfg containerConfigs) (string, error) {
	ctx := context.Background()
	container, err := e.cli.ContainerCreate(ctx, &cfg.config, &cfg.hostConfig, &cfg.networkConfig, name)
	if err != nil {
		e.log.WithError(err).Warnf("failed to create container (%s)", name)
		return "", err
	}
	err = e.cli.ContainerStart(ctx, container.ID, types.ContainerStartOptions{})
	if err != nil {
		e.log.WithError(err).Warnf("failed to start container (%s:%s)", container.ID[:12], name)
		return "", err
	}
	e.log.Debugf("container (%s:%s) started", container.ID[:12], name)
	return container.ID, nil
}

func (e *dockerEngine) restartContainer(cid string) error {
	ctx := context.Background()
	err := e.cli.ContainerRestart(ctx, cid, &e.grace)
	if err != nil {
		e.log.WithError(err).Warnf("failed to restart container (%s)", cid[:12])
	} else {
		e.log.Debugf("container (%s) restarted", cid[:12])
	}
	return err
}

func (e *dockerEngine) waitContainer(cid string) error {
	t := time.Now()
	ctx := context.Background()
	statusChan, errChan := e.cli.ContainerWait(ctx, cid, container.WaitConditionNotRunning)
	select {
	case err := <-errChan:
		e.logsContainer(cid, t)
		e.log.WithError(err).Warnf("failed to wait container (%s)", cid[:12])
		return err
	case status := <-statusChan:
		e.logsContainer(cid, t)
		e.log.Debugf("container (%s) exit status: %v", cid[:12], status)
		if status.Error != nil {
			return fmt.Errorf(status.Error.Message)
		} else if status.StatusCode != 0 {
			return fmt.Errorf("container exit code: %d", status.StatusCode)
		} else {
			return nil
		}
	}
}

func (e *dockerEngine) stopContainer(cid string) error {
	e.log.Debugf("container (%s) is stopping", cid[:12])

	ctx := context.Background()
	err := e.cli.ContainerStop(ctx, cid, &e.grace)
	if err != nil {
		e.log.WithError(err).Warnf("failed to stop container (%s)", cid[:12])
		return err
	}
	statusChan, errChan := e.cli.ContainerWait(ctx, cid, container.WaitConditionNotRunning)
	select {
	case <-time.After(e.grace):
		// e.cli.ContainerKill(ctx, cid, "9")
		e.log.Warnf("timed out to wait container (%s)", cid[:12])
		return fmt.Errorf("timed out to wait container (%s)", cid[:12])
	case err := <-errChan:
		e.log.WithError(err).Warnf("failed to wait container (%s)", cid[:12])
		return err
	case status := <-statusChan:
		e.log.Debugf("container (%s) exit status: %v", cid[:12], status)
		return nil
	}
}

func (e *dockerEngine) removeContainer(cid string) error {
	ctx := context.Background()
	err := e.cli.ContainerRemove(ctx, cid, types.ContainerRemoveOptions{Force: true})
	if err != nil {
		e.log.WithError(err).Warnf("failed to remove container (%s)", cid[:12])
	} else {
		e.log.Debugf("container (%s) removed", cid[:12])
	}
	return err
}

func (e *dockerEngine) removeContainerByName(name string) error {
	ctx := context.Background()
	args := filters.NewArgs()
	args.Add("name", name)
	containers, err := e.cli.ContainerList(ctx, types.ContainerListOptions{Filters: args, All: true})
	if err != nil {
		e.log.WithError(err).Warnf("failed to list container (%s)", name)
		return err
	}
	if len(containers) == 0 {
		return nil
	}
	err = e.cli.ContainerRemove(ctx, containers[0].ID, types.ContainerRemoveOptions{Force: true})
	if err != nil {
		e.log.WithError(err).Warnf("failed to remove container (%s:%s)", containers[0].ID[:12], name)
	} else {
		e.log.Debugf("container (%s:%s) removed", containers[0].ID[:12], name)
	}
	return err
}

func (e *dockerEngine) statsContainer(cid string) engine.PartialStats {
	t := time.Now().UTC()
	ctx := context.Background()
	sresp, err := e.cli.ContainerStats(ctx, cid, false)
	if err != nil {
		e.log.WithError(err).Warnf("failed to stats container (%s)", cid[:12])
		return engine.PartialStats{"error": err.Error()}
	}
	defer sresp.Body.Close()
	var tstats types.Stats
	err = json.NewDecoder(sresp.Body).Decode(&tstats)
	if err != nil {
		e.log.WithError(err).Warnf("failed to read stats response of container (%s)", cid[:12])
		return engine.PartialStats{"error": err.Error()}
	}

	if tstats.Read.IsZero() || tstats.PreRead.IsZero() {
		return nil
	}

	var cpuPercent = 0.0
	if runtime.GOOS != "windows" {
		cpuDelta := float64(tstats.CPUStats.CPUUsage.TotalUsage - tstats.PreCPUStats.CPUUsage.TotalUsage)
		systemDelta := float64(tstats.CPUStats.SystemUsage - tstats.PreCPUStats.SystemUsage)
		cpuPercent = (cpuDelta / systemDelta) * float64(len(tstats.CPUStats.CPUUsage.PercpuUsage))
	} else {
		possIntervals := uint64(tstats.Read.Sub(tstats.PreRead).Nanoseconds()) / 100 * uint64(tstats.NumProcs)
		intervalsUsed := tstats.CPUStats.CPUUsage.TotalUsage - tstats.PreCPUStats.CPUUsage.TotalUsage
		if possIntervals > 0 {
			cpuPercent = float64(intervalsUsed) / float64(possIntervals)
		}
	}

	UsedPercent := 0.0
	if tstats.MemoryStats.Limit > 0 {
		UsedPercent = float64(tstats.MemoryStats.Usage) / float64(tstats.MemoryStats.Limit)
	}

	return engine.PartialStats{
		"cpu_stats": utils.CPUInfo{
			Time:        t,
			UsedPercent: cpuPercent,
		},
		"mem_stats": utils.MemInfo{
			Time:        t,
			Total:       tstats.MemoryStats.Limit,
			Used:        tstats.MemoryStats.Usage,
			UsedPercent: UsedPercent,
		},
	}
}

func (e *dockerEngine) logsContainer(cid string, since time.Time) error {
	ctx := context.Background()
	r, err := e.cli.ContainerLogs(ctx, cid, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Since:      since.Format("2006-01-02T15:04:05"),
	})
	if err != nil {
		e.log.WithError(err).Warnf("failed to log container (%s)", cid[:12])
		return err
	}
	defer r.Close()
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) <= 8 {
			continue
		}
		switch stdcopy.StdType(line[0]) {
		case stdcopy.Stderr:
			e.log.Errorf("container (%s) %s", cid[:12], string(line[8:]))
		case stdcopy.Stdin, stdcopy.Stdout:
			e.log.Infof("container (%s) %s", cid[:12], string(line[8:]))
		default:
			e.log.Infof("container (%s) %s", cid[:12], string(line))
		}
	}
	return nil
}
