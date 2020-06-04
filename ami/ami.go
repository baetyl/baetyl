package ami

import (
	"io"
	"os"
	"sync"

	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-go/log"
	specv1 "github.com/baetyl/baetyl-go/spec/v1"
	"github.com/pkg/errors"
)

//go:generate mockgen -destination=../mock/ami.go -package=mock github.com/baetyl/baetyl-core/ami AMI

const (
	Kubernetes = "kubernetes"
)

var mu sync.Mutex
var amiNews = map[string]New{}
var amiImpls = map[string]AMI{}

type New func(cfg config.EngineConfig) (AMI, error)

// AMI app model interfaces
type AMI interface {
	CollectNodeInfo() (*specv1.NodeInfo, error)
	CollectNodeStats() (*specv1.NodeStatus, error)
	CollectAppStatus(string) ([]specv1.AppStatus, error)
	DeleteApplication(string, string) error
	ApplyApplication(string, specv1.Application, []string) error
	ApplyConfigurations(string, map[string]specv1.Configuration) error
	ApplySecrets(string, map[string]specv1.Secret) error
	FetchLog(namespace, service string, tailLines, sinceSeconds int64) (io.ReadCloser, error)
}

func NewAMI(cfg config.EngineConfig) (AMI, error) {
	name := cfg.Kind
	mu.Lock()
	defer mu.Unlock()
	if ami, ok := amiImpls[name]; ok {
		return ami, nil
	}
	amiNew, ok := amiNews[name]
	if !ok {
		return nil, errors.Wrapf(os.ErrInvalid, "ami (%s) not exists", name)
	}
	ami, err := amiNew(cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create ami (%s)", name)
	}
	amiImpls[name] = ami
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
