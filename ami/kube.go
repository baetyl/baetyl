package ami

import (
	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/shadow"
	bh "github.com/timshannon/bolthold"
	"os"
)

type kubeModel struct {
	cli      *Client
	store    *bh.Store
	shadow   *shadow.Shadow
	nodeName string
}

// TODO: move store and shadow to engine. kubemodel only implement the interfaces of omi
func NewKubeModel(cfg config.KubernetesConfig, sto *bh.Store, sha *shadow.Shadow) (Model, error) {
	cli, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}
	nodeName := os.Getenv("NODE_NAME")
	return &kubeModel{cli: cli, store: sto, shadow: sha, nodeName: nodeName}, nil
}
