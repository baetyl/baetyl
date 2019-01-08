package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/baidu/openedge/module/config"
	"github.com/baidu/openedge/module/logger"
	"github.com/baidu/openedge/module/master"
	"github.com/baidu/openedge/module/utils"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

// DockerSpec for docker
type DockerSpec struct {
	context       *Context
	module        *config.Module
	client        *client.Client
	config        *container.Config
	hostConfig    *container.HostConfig
	networkConfig *network.NetworkingConfig
}

// DockerContainer docker container to run and retry
type DockerContainer struct {
	spec *DockerSpec
	cid  string // container id
	tomb utils.Tomb
	log  *logger.Entry
}

// NewDockerContainer create a new docker container
func NewDockerContainer(s *DockerSpec) Worker {
	return &DockerContainer{
		spec: s,
		log:  logger.WithFields("module", s.module.UniqueName()),
	}
}

// UniqueName unique name of worker
func (w *DockerContainer) UniqueName() string {
	return w.spec.module.UniqueName()
}

// Policy returns restart policy
func (w *DockerContainer) Policy() config.Policy {
	return w.spec.module.Restart
}

// Start starts container
func (w *DockerContainer) Start(supervising func(Worker) error) error {
	err := w.startContainer()
	if err != nil {
		return err
	}
	err = w.tomb.Go(func() error {
		return supervising(w)
	})
	return err
}

// Restart restarts container
func (w *DockerContainer) Restart() error {
	return w.restartContainer()
}

// Stop stops container with a gracetime
func (w *DockerContainer) Stop() error {
	if !w.tomb.Alive() {
		w.log.Debugf("container already stopped")
		return nil
	}
	w.tomb.Kill(nil)
	err := w.stopContainer()
	if err != nil {
		return err
	}
	return w.tomb.Wait()
}

// Wait waits until container is stopped
func (w *DockerContainer) Wait(c chan<- error) {
	defer w.log.Infof("container stopped")

	ctx := context.Background()
	statusChan, errChan := w.spec.client.ContainerWait(ctx, w.cid, container.WaitConditionNotRunning)
	select {
	case err := <-errChan:
		w.log.WithError(err).Warnln("failed to wait container")
		c <- err
	case status := <-statusChan:
		w.log.Infof("container exited: %v", status)
		c <- fmt.Errorf("container exited: %v", status)
	}
}

// Dying returns the channel that can be used to wait until container is stopped
func (w *DockerContainer) Dying() <-chan struct{} {
	return w.tomb.Dying()
}

// Stats returns the stats of docker container
func (w *DockerContainer) Stats() (*master.ModuleStats, error) {
	ctx := context.Background()
	iresp, err := w.spec.client.ContainerInspect(ctx, w.cid)
	if err != nil {
		return nil, err
	}
	sresp, err := w.spec.client.ContainerStats(ctx, w.cid, false)
	if err != nil {
		return nil, err
	}
	defer sresp.Body.Close()
	data, err := ioutil.ReadAll(sresp.Body)
	if err != nil {
		return nil, err
	}
	var tstats types.Stats
	err = json.Unmarshal(data, &tstats)
	if err != nil {
		return nil, err
	}
	return &master.ModuleStats{
		Stats:      tstats,
		Status:     iresp.State.Status,
		StartedAt:  iresp.State.StartedAt,
		FinishedAt: iresp.State.FinishedAt,
	}, nil
}

func (w *DockerContainer) startContainer() error {
	ctx := context.Background()
	container, err := w.spec.client.ContainerCreate(ctx, w.spec.config, w.spec.hostConfig, w.spec.networkConfig, w.spec.module.UniqueName())
	if err != nil {
		w.log.WithError(err).Warnln("failed to create container")
		// stop, remove and retry
		w.removeContainerByName()
		container, err = w.spec.client.ContainerCreate(ctx, w.spec.config, w.spec.hostConfig, w.spec.networkConfig, w.spec.module.UniqueName())
		if err != nil {
			w.log.WithError(err).Warnln("failed to create container again")
			return err
		}
	}
	w.cid = container.ID
	w.log = w.log.WithFields("cid", container.ID[:12])
	err = w.spec.client.ContainerStart(ctx, w.cid, types.ContainerStartOptions{})
	if err != nil {
		w.log.WithError(err).Warnln("failed to start container")
		return err
	}
	w.log.Infof("container started")
	return nil
}

func (w *DockerContainer) restartContainer() error {
	ctx := context.Background()
	err := w.spec.client.ContainerRestart(ctx, w.cid, &w.spec.context.Grace)
	if err != nil {
		w.log.Warnf("failed to restart container")
	}
	return err
}

func (w *DockerContainer) stopContainer() error {
	if w.cid == "" {
		return nil
	}
	ctx := context.Background()
	err := w.spec.client.ContainerStop(ctx, w.cid, &w.spec.context.Grace)
	if err != nil {
		w.log.Errorf("failed to stop container")
		return err
	}
	err = w.spec.client.ContainerRemove(ctx, w.cid, types.ContainerRemoveOptions{Force: true})
	if err != nil {
		w.log.Warnf("failed to remove container")
	} else {
		w.log.Infof("container removed")
	}
	return nil
}

func (w *DockerContainer) removeContainerByName() {
	ctx := context.Background()
	args := filters.NewArgs()
	args.Add("name", w.spec.module.UniqueName())
	containers, err := w.spec.client.ContainerList(ctx, types.ContainerListOptions{Filters: args, All: true})
	if err != nil {
		w.log.WithError(err).Warnf("failed to list containers (%s)", w.spec.module.UniqueName())
	}
	for _, c := range containers {
		err := w.spec.client.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{Force: true})
		if err != nil {
			w.log.WithError(err).Warnf("failed to remove old container (%s:%v)", c.ID[:12], c.Names)
		} else {
			w.log.Infof("old container (%s:%v) removed", c.ID[:12], c.Names)
		}
	}
}
