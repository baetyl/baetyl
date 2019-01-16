package engine

import (
	"context"

	openedge "github.com/baidu/openedge/api/go"
	"github.com/baidu/openedge/master/engine"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

// NAME ot docker engine
const NAME = "docker"
const defaultNetworkName = "openedge"

func init() {
	engine.Factories()[NAME] = New
}

// New docker engine
func New(wdir string) (engine.Engine, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	d := &docker{
		client: cli,
		wdir:   wdir,
		log:    openedge.WithField("mode", "docker"),
	}
	err = d.initNetwork()
	if err != nil {
		return nil, err
	}
	return d, nil
}

type docker struct {
	client  *client.Client
	wdir    string
	network string
	log     openedge.Logger
}

func (d *docker) Name() string {
	return NAME
}

func (d *docker) Close() error {
	return d.client.Close()
}

func (d *docker) initNetwork() error {
	context := context.Background()
	args := filters.NewArgs()
	args.Add("driver", "bridge")
	args.Add("type", "custom")
	args.Add("name", defaultNetworkName)
	nws, err := d.client.NetworkList(context, types.NetworkListOptions{Filters: args})
	if err != nil {
		return err
	}
	if len(nws) > 0 {
		d.network = nws[0].ID
		d.log.Infof("network (%s:openedge) exists", d.network[:12])
		return nil
	}
	nw, err := d.client.NetworkCreate(context, defaultNetworkName, types.NetworkCreate{Driver: "bridge", Scope: "local"})
	if err != nil {
		return err
	}
	if nw.Warning != "" {
		d.log.Warnf(nw.Warning)
	}
	d.network = nw.ID
	d.log.Infof("network (%s:openedge) created", d.network[:12])
	return nil
}

/*
// Prepare prepares images
func (e *DockerEngine) Prepare(image string) error {
	out, err := e.client.ImagePull(context.Background(), image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer out.Close()
	io.Copy(os.Stdout, out)
	return nil
}
*/
