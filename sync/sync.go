package sync

import (
	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/event"
	"github.com/baetyl/baetyl-core/models"
	"github.com/baetyl/baetyl-core/shadow"
	"github.com/baetyl/baetyl-go/http"
	"github.com/baetyl/baetyl-go/link"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/mqtt"
	"github.com/baetyl/baetyl-go/utils"
	bh "github.com/timshannon/bolthold"
	appv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
)

//go:generate mockgen -destination=../mock/sync.go -package=mock github.com/baetyl/baetyl-core/sync Sync

type node struct {
	Name      string
	Namespace string
}

type batch struct {
	Name      string
	Namespace string
}

type Sync interface {
	Close()
	Report(msg link.Message) error
	ProcessDelta(msg link.Message) error
	ProcessVolumes(volumes []models.Volume, cfgs map[string]*models.Configuration) error
	ProcessConfiguration(volume *models.Volume, cfg *models.Configuration) error
	ProcessApplication(app *models.Application) error
}

func NewSync(cfg config.SyncConfig, sto *bh.Store, sha *shadow.Shadow, cent *event.Center) (Sync, error) {
	httpOps, err := cfg.Cloud.HTTP.ToClientOptions()
	if err != nil {
		return nil, err
	}
	httpCli := http.NewClient(*httpOps)
	s := &sync{
		cent:   cent,
		cfg:    cfg,
		store:  sto,
		shadow: sha,
		http:   httpCli,
		node: &node{
			Name:      cfg.Node.Name,
			Namespace: cfg.Node.Namespace,
		},
		log: log.With(log.Any("core", "sync")),
	}
	return s, nil
}

type sync struct {
	cent   *event.Center
	log    *log.Logger
	cfg    config.SyncConfig
	tomb   utils.Tomb
	impl   appv1.DeploymentInterface
	http   *http.Client
	mqtt   *mqtt.Client
	node   *node
	store  *bh.Store
	shadow *shadow.Shadow
}

func (s *sync) Close() {
	s.tomb.Kill(nil)
	s.tomb.Wait()
}
