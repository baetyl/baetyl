package engine

import (
	"context"

	"github.com/baidu/openedge/module"
	"github.com/baidu/openedge/utils"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/juju/errors"
)

// DockerSpec for docker
type DockerSpec struct {
	Spec
	Client        *client.Client
	Config        *container.Config
	HostConfig    *container.HostConfig
	NetworkConfig *network.NetworkingConfig
}

// DockerContainer docker container to run and retry
type DockerContainer struct {
	spec *DockerSpec
	cid  string // container id
	tomb utils.Tomb
}

// NewDockerContainer create a new docker container
func NewDockerContainer(s *DockerSpec) Worker {
	return &DockerContainer{
		spec: s,
	}
}

// Policy returns name
func (w *DockerContainer) Name() string {
	return w.spec.Name
}

// Policy returns restart policy
func (w *DockerContainer) Policy() module.Policy {
	return w.spec.Restart
}

// Start starts container
func (w *DockerContainer) Start(supervising func(Worker) error) error {
	err := w.startContainer()
	if err != nil {
		return errors.Trace(err)
	}
	err = w.tomb.Go(func() error {
		return supervising(w)
	})
	return errors.Trace(err)
}

// Restart restarts container
func (w *DockerContainer) Restart() error {
	return errors.Trace(w.restartContainer())
}

// Stop stops container with a gracetime
func (w *DockerContainer) Stop() error {
	if !w.tomb.Alive() {
		w.spec.Logger.Debug("Container already stopped")
		return nil
	}
	w.tomb.Kill(nil)
	err := w.stopContainer()
	if err != nil {
		return errors.Trace(err)
	}
	return errors.Trace(w.tomb.Wait())
}

// Wait waits until container is stopped
func (w *DockerContainer) Wait(c chan<- error) {
	defer w.spec.Logger.Info("Container stopped")
	ctx := context.Background()
	statusChan, errChan := w.spec.Client.ContainerWait(ctx, w.cid, container.WaitConditionNotRunning)
	select {
	case err := <-errChan:
		w.spec.Logger.WithError(err).Warnln("Failed to wait container")
		c <- err
	case status := <-statusChan:
		w.spec.Logger.Infof("Container exited: %v", status)
		c <- errors.Errorf("Container exited: %v", status)
	}
}

// Dying returns the channel that can be used to wait until container is stopped
func (w *DockerContainer) Dying() <-chan struct{} {
	return w.tomb.Dying()
}

func (w *DockerContainer) startContainer() error {
	ctx := context.Background()
	container, err := w.spec.Client.ContainerCreate(ctx, w.spec.Config, w.spec.HostConfig, w.spec.NetworkConfig, w.spec.Name)
	if err != nil {
		w.spec.Logger.WithError(err).Warnln("Failed to create container")
		// stop, remove and retry
		w.removeContainerByName()
		container, err = w.spec.Client.ContainerCreate(ctx, w.spec.Config, w.spec.HostConfig, w.spec.NetworkConfig, w.spec.Name)
		if err != nil {
			w.spec.Logger.WithError(err).Warnln("Failed to create container again")
			return errors.Trace(err)
		}
	}
	w.cid = container.ID
	w.spec.Logger = w.spec.Logger.WithField("cid", container.ID[:12])
	err = w.spec.Client.ContainerStart(ctx, w.cid, types.ContainerStartOptions{})
	if err != nil {
		w.spec.Logger.WithError(err).Warnln("Failed to start container")
		return errors.Trace(err)
	}
	w.spec.Logger.Infof("Container started")
	return nil
}

func (w *DockerContainer) restartContainer() error {
	ctx := context.Background()
	err := w.spec.Client.ContainerRestart(ctx, w.cid, &w.spec.Grace)
	if err != nil {
		w.spec.Logger.Warnf("Failed to restart container")
	}
	return errors.Trace(err)
}

func (w *DockerContainer) stopContainer() error {
	if w.cid == "" {
		return nil
	}
	ctx := context.Background()
	err := w.spec.Client.ContainerStop(ctx, w.cid, &w.spec.Grace)
	if err != nil {
		w.spec.Logger.Error("Failed to stop container")
		return errors.Trace(err)
	}
	err = w.spec.Client.ContainerRemove(ctx, w.cid, types.ContainerRemoveOptions{Force: true})
	if err != nil {
		w.spec.Logger.Warnf("Failed to remove container")
	} else {
		w.spec.Logger.Info("Container removed")
	}
	return nil
}

func (w *DockerContainer) removeContainerByName() {
	ctx := context.Background()
	args := filters.NewArgs()
	args.Add("name", w.spec.Name)
	containers, err := w.spec.Client.ContainerList(ctx, types.ContainerListOptions{Filters: args, All: true})
	if err != nil {
		w.spec.Logger.WithError(err).Warnf("Failed to list containers (%s)", w.spec.Name)
	}
	for _, c := range containers {
		err := w.spec.Client.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{Force: true})
		if err != nil {
			w.spec.Logger.WithError(err).Warnf("Failed to remove old container (%s:%v)", c.ID[:12], c.Names)
		} else {
			w.spec.Logger.Infof("Old container (%s:%v) removed", c.ID[:12], c.Names)
		}
	}
}
