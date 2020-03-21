package ami

import (
	"os"

	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/shadow"
	"github.com/baetyl/baetyl-go/log"
	bh "github.com/timshannon/bolthold"
)

type kubeModel struct {
	cli      *Client
	store    *bh.Store
	shadow   *shadow.Shadow
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
	return &kubeModel{
		cli:      cli,
		store:    sto,
		nodeName: nodeName,
		log:      log.With(log.Any("ami", "kube")),
	}, nil
}
