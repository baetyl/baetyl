package docker

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
)

const defaultNetworkName = "openedge"

type containerConfigs struct {
	config        container.Config
	hostConfig    container.HostConfig
	networkConfig network.NetworkingConfig
}

func (e *dockerEngine) initNetwork() error {
	ctx := context.Background()
	args := filters.NewArgs()
	args.Add("driver", "bridge")
	args.Add("type", "custom")
	args.Add("name", defaultNetworkName)
	nws, err := e.cli.NetworkList(ctx, types.NetworkListOptions{Filters: args})
	if err != nil {
		e.log.WithError(err).Errorf("failed to list network (%s)", defaultNetworkName)
		return err
	}
	if len(nws) > 0 {
		e.nid = nws[0].ID
		e.log.Debugf("network (%s:%s) exists", e.nid[:12], defaultNetworkName)
		return nil
	}
	nw, err := e.cli.NetworkCreate(ctx, defaultNetworkName, types.NetworkCreate{Driver: "bridge", Scope: "local"})
	if err != nil {
		e.log.WithError(err).Errorf("failed to create network (%s)", defaultNetworkName)
		return err
	}
	if nw.Warning != "" {
		e.log.Warnf(nw.Warning)
	}
	e.nid = nw.ID
	e.log.Debugf("network (%s:%s) created", e.nid[:12], defaultNetworkName)
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
	ctx := context.Background()
	statusChan, errChan := e.cli.ContainerWait(ctx, cid, container.WaitConditionNotRunning)
	select {
	case err := <-errChan:
		e.log.WithError(err).Warnf("failed to wait container (%s)", cid[:12])
		return err
	case status := <-statusChan:
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
	e.log.Debugf("to stop container (%s)", cid[:12])

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
