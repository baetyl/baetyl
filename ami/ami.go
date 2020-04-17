package ami

import (
	"github.com/baetyl/baetyl-core/config"
	specv1 "github.com/baetyl/baetyl-go/spec/v1"
	bh "github.com/timshannon/bolthold"
	"io"
	"os"
)

//go:generate mockgen -destination=../mock/ami.go -package=mock github.com/baetyl/baetyl-core/ami AMI

// AMI app model interfaces
type AMI interface {
	Collect(namespace string) (specv1.Report, error)
	Apply(namespace string, appInfos []specv1.AppInfo, condition string) error
	FetchLog(namespace, service string, tailLines, sinceSeconds int64) (io.ReadCloser, error)
}

const (
	Kubernetes = "kubernetes"
)

func GenAMI(cfg config.EngineConfig, sto *bh.Store) (AMI, error) {
	switch cfg.Kind {
	case Kubernetes:
		return newKubeImpl(cfg.Kubernetes, sto)
	default:
		return nil, os.ErrInvalid
	}
}
