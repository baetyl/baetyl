package engine_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
)

func TestDockerClient(t *testing.T) {
	cli, err := client.NewEnvClient()
	ctx := context.Background()
	inspectResp, err := cli.ContainerInspect(ctx, "809876f19c02")
	fmt.Println(err)
	if err == nil {
		fmt.Println(inspectResp.State.Status)
	}
	resp, err := cli.ContainerStats(ctx, "809876f19c02", false)
	assert.NoError(t, err)
	data, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	fmt.Println(string(data))
	var stats types.Stats
	err = json.Unmarshal(data, &stats)
	assert.NoError(t, err)
	fmt.Println(stats)
}
