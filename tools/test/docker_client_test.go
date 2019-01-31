package test

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
)

func TestInspect(t *testing.T) {
	cli, err := client.NewEnvClient()
	assert.NoError(t, err)

	id := "d2cc22d7fa01"
	ctx := context.Background()
	iresp, err := cli.ContainerInspect(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, "Status:", iresp.State.Status)
}

func TestWait(t *testing.T) {
	cli, err := client.NewEnvClient()
	assert.NoError(t, err)

	id := "d2cc22d7fa01"
	ctx := context.Background()
	statusChan, errChan := cli.ContainerWait(ctx, id, container.WaitConditionNotRunning)
	select {
	case err := <-errChan:
		assert.NoError(t, err)
	case status := <-statusChan:
		assert.Equal(t, "Status:", status)
	}
}

func TestStart(t *testing.T) {
	timeout := 10
	containerConfig := &container.Config{
		Image:       "openedge-remote-mqtt",
		Labels:      map[string]string{"openedge": "service"},
		StopTimeout: &timeout,
	}
	containerHostConfig := &container.HostConfig{
		RestartPolicy: container.RestartPolicy{
			Name: "always",
			// MaximumRetryCount: 2,
		},
	}

	cli, err := client.NewEnvClient()
	assert.NoError(t, err)

	ctx := context.Background()
	container, err := cli.ContainerCreate(ctx, containerConfig, containerHostConfig, nil, "")
	assert.NoError(t, err)

	err = cli.ContainerStart(ctx, container.ID, types.ContainerStartOptions{})
	assert.NoError(t, err)
}
