package docker

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
)

func Test_dockerEngine_pullImage(t *testing.T) {
	cli, err := client.NewEnvClient()
	name := "eclipse-mosquitto"
	out, err := cli.ImagePull(context.Background(), name, types.ImagePullOptions{})
	assert.NoError(t, err)
	defer out.Close()
	io.Copy(os.Stdout, out)
}
