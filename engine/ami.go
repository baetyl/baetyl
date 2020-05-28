package engine

import (
	"github.com/baetyl/baetyl-core/ami"
	"github.com/baetyl/baetyl-core/config"
	specv1 "github.com/baetyl/baetyl-go/spec/v1"
	bh "github.com/timshannon/bolthold"
	"io"
	"os"
)

//go:generate mockgen -destination=../mock/ami.go -package=mock github.com/baetyl/baetyl-core/engine AMI

const (
	Kubernetes = "kubernetes"
)

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

func GenAMI(cfg config.EngineConfig, sto *bh.Store) (AMI, error) {
	switch cfg.Kind {
	case Kubernetes:
		return ami.NewKubeImpl(cfg.Kubernetes, sto)
	default:
		return nil, os.ErrInvalid
	}
}
