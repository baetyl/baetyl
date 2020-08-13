package kube

import (
	"os"

	"github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl/ami"
	"github.com/baetyl/baetyl/config"
	bh "github.com/timshannon/bolthold"
)

type kubeImpl struct {
	knn   string // kube node name
	cli   *client
	store *bh.Store
	conf  *config.KubernetesConfig
	log   *log.Logger
}

func init() {
	ami.Register("kube", newKubeImpl)
	ami.Register("kubernetes", newKubeImpl)
}

func newKubeImpl(cfg config.AmiConfig) (ami.AMI, error) {
	cli, err := newClient(cfg.Kubernetes)
	if err != nil {
		return nil, err
	}
	knn := os.Getenv(KubeNodeName)
	model := &kubeImpl{
		knn:  knn,
		cli:  cli,
		conf: &cfg.Kubernetes,
		log:  log.With(log.Any("ami", "kube")),
	}
	return model, nil
}

func (k *kubeImpl) ApplyApp(s string, application specv1.Application, m map[string]specv1.Configuration, m2 map[string]specv1.Secret) error {
	panic("implement me")
}

func (k *kubeImpl) StatsApp(s string) ([]specv1.AppStats, error) {
	panic("implement me")
}

func (k *kubeImpl) DeleteApp(s string, s2 string) error {
	panic("implement me")
}
