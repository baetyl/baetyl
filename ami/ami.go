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

var mu sync.Mutex
var amiNews = map[string]New{}
var amiImpls = map[string]AMI{}

type New func(cfg config.AmiConfig) (AMI, error)

// AMI app model interfaces
type AMI interface {
	// node
	CollectNodeInfo() (*specv1.NodeInfo, error)
	CollectNodeStats() (*specv1.NodeStats, error)

	// app
	ApplyApp(string, specv1.Application, map[string]specv1.Configuration, map[string]specv1.Secret) error
	DeleteApp(string, string) error
	StatsApps(string) ([]specv1.AppStats, error)

	// TODO: update
	FetchLog(namespace, service string, tailLines, sinceSeconds int64) (io.ReadCloser, error)

	RemoteCommand(option DebugOptions, stdin io.Reader, stdout, stderr io.Writer) error
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
