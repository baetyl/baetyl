package ami

import (
	"io"
	"os"
	"sync"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl/config"
)

//go:generate mockgen -destination=../mock/ami.go -package=mock github.com/baetyl/baetyl/ami AMI

const (
	Kubernetes = "kubernetes"
)

var mu sync.Mutex
var amiNews = map[string]New{}
var amiImpls = map[string]AMI{}

type New func(cfg config.AmiConfig) (AMI, error)

// AMI app model interfaces
type AMI interface {
	CollectNodeInfo() (*specv1.NodeInfo, error)
	CollectNodeStats() (*specv1.NodeStats, error)
	CollectAppStats(string) ([]specv1.AppStats, error)
	DeleteApplication(string, string) error
	ApplyApplication(string, specv1.Application, []string) error
	ApplyConfigurations(string, map[string]specv1.Configuration) error
	ApplySecrets(string, map[string]specv1.Secret) error
	FetchLog(namespace, service string, tailLines, sinceSeconds int64) (io.ReadCloser, error)
}

func NewAMI(cfg config.AmiConfig) (AMI, error) {
	name := cfg.Kind
	mu.Lock()
	defer mu.Unlock()
	if ami, ok := amiImpls[name]; ok {
		return ami, nil
	}
	amiNew, ok := amiNews[name]
	if !ok {
		return nil, errors.Trace(os.ErrInvalid)
	}
	ami, err := amiNew(cfg)
	if err != nil {
		return nil, errors.Trace(err)
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
