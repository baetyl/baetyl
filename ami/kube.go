package ami

import (
	"os"

	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-go/log"
	bh "github.com/timshannon/bolthold"
)

type kubeModel struct {
	cli      *Client
	store    *bh.Store
	nodeName string
	log      *log.Logger
}

// TODO: move store and shadow to engine. kubemodel only implement the interfaces of omi
func NewKubeModel(cfg config.KubernetesConfig, sto *bh.Store) (AMI, error) {
	cli, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}
	nodeName := os.Getenv("NODE_NAME")
	model := &kubeModel{
		cli:      cli,
		store:    sto,
		nodeName: nodeName,
		log:      log.With(log.Any("ami", "kube")),
	}
	return model, nil
}
