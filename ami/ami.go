package ami

import (
	"io"
	"os"
	"sync"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/baetyl/baetyl/v2/config"
)

//go:generate mockgen -destination=../mock/ami.go -package=mock -source=ami.go AMI

const (
	BaetylStatsExtension = "baetyl_stats_extension"
	BaetylSetAffinity    = "baetyl_set_affinity"
)

var mu sync.Mutex
var amiNews = map[string]New{}
var amiImpls = map[string]AMI{}
var Hooks = map[string]interface{}{}

type CollectStatsExtFunc func() (map[string]interface{}, error)
type SetAffinityFunc func(string) (*corev1.Affinity, error)

type New func(cfg config.AmiConfig) (AMI, error)

type Pipe struct {
	InReader  *io.PipeReader
	InWriter  *io.PipeWriter
	OutReader *io.PipeReader
	OutWriter *io.PipeWriter
}

// AMI app model interfaces
type AMI interface {
	// node
	// TODO remove GetMasterNodeName
	GetMasterNodeName() string
	CollectNodeInfo() (map[string]interface{}, error)
	CollectNodeStats() (map[string]interface{}, error)

	// app
	ApplyApp(string, specv1.Application, map[string]specv1.Configuration, map[string]specv1.Secret) error
	DeleteApp(string, string) error
	StatsApps(string) ([]specv1.AppStats, error)

	// TODO: update
	FetchLog(namespace, service string, tailLines, sinceSeconds int64) (io.ReadCloser, error)

	// RemoteCommand remote debug
	RemoteCommand(option DebugOptions, pipe Pipe) error

	UpdateNodeLabels(string, map[string]string) error
}

type DebugOptions struct {
	Namespace string
	Name      string
	Container string
	Command   []string
}

func NewAMI(mode string, cfg config.AmiConfig) (AMI, error) {
	mu.Lock()
	defer mu.Unlock()
	if ami, ok := amiImpls[mode]; ok {
		return ami, nil
	}
	amiNew, ok := amiNews[mode]
	if !ok {
		return nil, errors.Trace(os.ErrInvalid)
	}
	ami, err := amiNew(cfg)
	if err != nil {
		return nil, errors.Trace(err)
	}
	amiImpls[mode] = ami
	return ami, nil
}

func Register(name string, n New) {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := amiNews[name]; ok {
		log.L().Warn("ami generator already exists, skip", log.Any("generator", name))
		return
	}
	log.L().Debug("ami generator registered", log.Any("generator", name))
	amiNews[name] = n
}
