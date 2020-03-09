package agent

import (
	"github.com/baetyl/baetyl-core/common"
	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/models"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/utils"
	"github.com/baetyl/baetyl/protocol/http"
	appv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"os"
)

//go:generate mockgen -destination=../mock/agent/agent.go -package=plugin github.com/baetyl/baetyl-core/agent Agent

type node struct {
	Name      string
	Namespace string
}

type batch struct {
	Name      string
	Namespace string
}

type Agent interface {
	Start()
	Stop()
	Report()
	ProcessResource(le *EventLink) error
	ProcessVolumes(volumes []models.Volume, configs map[string]models.Configuration) error
	ProcessConfiguration(volume models.Volume, cfg models.Configuration) error
	ProcessApplication(app models.Application) error
}

func NewAgent(cfg config.AgentConfig, impl appv1.DeploymentInterface, log *log.Logger) (Agent, error) {
	httpCli, err := http.NewClient(*cfg.Remote.HTTP)
	if err != nil {
		return nil, err
	}
	a := &agent{
		log:     log,
		cfg:     cfg,
		impl:    impl,
		events:  make(chan *Event, 1),
		http:    httpCli,
	}
	return a, nil
}

type agent struct {
	log     *log.Logger
	cfg     config.AgentConfig
	tomb    utils.Tomb
	impl    appv1.DeploymentInterface
	events  chan *Event
	http    *http.Client
	batch   *batch
	node    *node
	shadow  *models.Shadow
}

func (a *agent) Start() {
	nodeName := os.Getenv(common.NodeName)
	nodeNamespace := os.Getenv(common.NodeNamespace)
	if nodeName != "" && nodeNamespace != "" {
		a.node = &node{
			Name:      nodeName,
			Namespace: nodeNamespace,
		}
	} else {
		batchName := os.Getenv(common.BatchName)
		batchNamespace := os.Getenv(common.BatchNamespace)
		a.batch = &batch{
			Name:      batchName,
			Namespace: batchNamespace,
		}
	}

	err := a.tomb.Go(a.reporting, a.processing)
	if err != nil {
		a.log.Error("failed to start report and process routine", log.Error(err))
		return
	}
}

func (a *agent) Stop() {
	a.tomb.Kill(nil)
	a.tomb.Wait()
}
