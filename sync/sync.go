package sync

import (
	"os"

	"github.com/baetyl/baetyl-core/common"
	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/models"
	"github.com/baetyl/baetyl-go/context"
	"github.com/baetyl/baetyl-go/http"
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
	Start()
	Stop()
	Report()
	ProcessResource(map[string]string) error
	ProcessVolumes(volumes []models.Volume, cfgs map[string]*models.Configuration) error
	ProcessConfiguration(volume *models.Volume, cfg *models.Configuration) error
	ProcessApplication(app *models.Application) error
}

func NewSync(ctx context.Context, cfg config.SyncConfig, store *bh.Store, log *log.Logger) (Sync, error) {
	httpOps, err := cfg.Cloud.HTTP.ToClientOptions()
	if err != nil {
		return nil, err
	}
	httpCli := http.NewClient(*httpOps)
	s := &sync{
		log:    log,
		cfg:    cfg,
		store:  store,
		events: make(chan *Event, 1),
		http:   httpCli,
	}
	return s, nil
}

type sync struct {
	log    *log.Logger
	cfg    config.SyncConfig
	tomb   utils.Tomb
	impl   appv1.DeploymentInterface
	events chan *Event
	http   *http.Client
	mqtt   *mqtt.Client
	batch  *batch
	node   *node
	store  *bh.Store
	shadow *models.Shadow
}

func (s *sync) Start() {
	nodeName := os.Getenv(common.NodeName)
	nodeNamespace := os.Getenv(common.NodeNamespace)
	if nodeName != "" && nodeNamespace != "" {
		s.node = &node{
			Name:      nodeName,
			Namespace: nodeNamespace,
		}
	} else {
		batchName := os.Getenv(common.BatchName)
		batchNamespace := os.Getenv(common.BatchNamespace)
		s.batch = &batch{
			Name:      batchName,
			Namespace: batchNamespace,
		}
	}

	err := s.tomb.Go(s.reporting, s.processing)
	if err != nil {
		s.log.Error("failed to start report and process routine", log.Error(err))
		return
	}
}

func (s *sync) Stop() {
	s.tomb.Kill(nil)
	s.tomb.Wait()
}
