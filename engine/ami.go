package engine

import (
	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-go/log"
	specv1 "github.com/baetyl/baetyl-go/spec/v1"
	"io"
	"os"
	"sync"
)

//go:generate mockgen -destination=../mock/ami.go -package=mock github.com/baetyl/baetyl-core/engine AMI

const (
	Kubernetes = "kubernetes"
)

var mu sync.Mutex

type New func(cfg config.EngineConfig) (AMI, error)

var amiNews = map[string]New{}
var amis = map[string]AMI{}

// AMI app model interfaces
type AMI interface {
	CollectNodeInfo() (*specv1.NodeInfo, error)
	CollectNodeStats() (*specv1.NodeStatus, error)
	CollectAppStatus(string) ([]specv1.AppStatus, error)
	DeleteApplication(string, specv1.Application) error
	ApplyApplication(string, specv1.Application, []string) error
	ApplyConfigurations(string, map[string]specv1.Configuration) error
	ApplySecrets(string, map[string]specv1.Secret) error
	FetchLog(namespace, service string, tailLines, sinceSeconds int64) (io.ReadCloser, error)
}

func GenAMI(cfg config.EngineConfig) (AMI, error) {
	name := cfg.Kind
	mu.Lock()
	defer mu.Unlock()
	if ami, ok := amis[name]; ok {
		return ami, nil
	}
	amiNew, ok := amiNews[name]
	if !ok {
		log.L().Error("ami generator not exists", log.Any("generator", name))
		return nil, os.ErrInvalid
	}
	ami, err := amiNew(cfg)
	if err != nil {
		log.L().Error("failed to generate ami", log.Any("generator", name))
		return nil, err
	}
	amis[name] = ami
	return ami, nil
}

func Register(name string, n New) {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := amiNews[name]; ok {
		log.L().Info("ami generator already exists, skip", log.Any("generator", name))
		return
	}
	log.L().Info("ami generator registered", log.Any("generator", name))
	amiNews[name] = n
}
