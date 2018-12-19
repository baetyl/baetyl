package engine

import (
	"context"
	"fmt"

	"github.com/baidu/openedge/module/config"
	"github.com/baidu/openedge/module/utils"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
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

// Name returns name
func (w *DockerContainer) Name() string {
	return w.spec.Name
}

// Policy returns restart policy
func (w *DockerContainer) Policy() config.Policy {
	return w.spec.Restart
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
		w.spec.Logger.Debugf("container already stopped")
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
	defer w.spec.Logger.Infof("container stopped")

	ctx := context.Background()
	statusChan, errChan := w.spec.Client.ContainerWait(ctx, w.cid, container.WaitConditionNotRunning)
	select {
	case err := <-errChan:
		w.spec.Logger.WithError(err).Warnln("failed to wait container")
		c <- err
	case status := <-statusChan:
		w.spec.Logger.Infof("container exited: %v", status)
		c <- fmt.Errorf("container exited: %v", status)
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
		w.spec.Logger.WithError(err).Warnln("failed to create container")
		// stop, remove and retry
		w.removeContainerByName()
		container, err = w.spec.Client.ContainerCreate(ctx, w.spec.Config, w.spec.HostConfig, w.spec.NetworkConfig, w.spec.Name)
		if err != nil {
			w.spec.Logger.WithError(err).Warnln("failed to create container again")
			return err
		}
	}
	w.cid = container.ID
	w.spec.Logger = w.spec.Logger.WithFields("cid", container.ID[:12])
	err = w.spec.Client.ContainerStart(ctx, w.cid, types.ContainerStartOptions{})
	if err != nil {
		w.spec.Logger.WithError(err).Warnln("failed to start container")
		return err
	}
	w.spec.Logger.Infof("container started")
	return nil
}

func (w *DockerContainer) restartContainer() error {
	ctx := context.Background()
	err := w.spec.Client.ContainerRestart(ctx, w.cid, &w.spec.Grace)
	if err != nil {
		w.spec.Logger.Warnf("failed to restart container")
	}
	return err
}

func (w *DockerContainer) stopContainer() error {
	if w.cid == "" {
		return nil
	}
	ctx := context.Background()
	err := w.spec.Client.ContainerStop(ctx, w.cid, &w.spec.Grace)
	if err != nil {
		w.spec.Logger.Errorf("failed to stop container")
		return err
	}
	err = w.spec.Client.ContainerRemove(ctx, w.cid, types.ContainerRemoveOptions{Force: true})
	if err != nil {
		w.spec.Logger.Warnf("failed to remove container")
	} else {
		w.spec.Logger.Infof("container removed")
	}
	return nil
}

func (w *DockerContainer) removeContainerByName() {
	ctx := context.Background()
	args := filters.NewArgs()
	args.Add("name", w.spec.Name)
	containers, err := w.spec.Client.ContainerList(ctx, types.ContainerListOptions{Filters: args, All: true})
	if err != nil {
		w.spec.Logger.WithError(err).Warnf("failed to list containers (%s)", w.spec.Name)
	}
	for _, c := range containers {
		err := w.spec.Client.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{Force: true})
		if err != nil {
			w.spec.Logger.WithError(err).Warnf("failed to remove old container (%s:%v)", c.ID[:12], c.Names)
		} else {
			w.spec.Logger.Infof("old container (%s:%v) removed", c.ID[:12], c.Names)
		}
	}
}
