package sync

import (
	"fmt"
	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl/plugin"
	"time"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/baetyl/baetyl/config"
	"github.com/baetyl/baetyl/node"
	bh "github.com/timshannon/bolthold"
	"k8s.io/apimachinery/pkg/util/rand"
)

var (
	// ErrSyncTLSConfigMissing certificate bidirectional authentication is required for connection with cloud
	ErrSyncTLSConfigMissing = fmt.Errorf("certificate bidirectional authentication is required for connection with cloud")
)

const (
	EnvKeyNodeName      = "BAETYL_NODE_NAME"
	EnvKeyNodeNamespace = "BAETYL_NODE_NAMESPACE"
)

const (
	RequestType   = "request-type"
	ReportRequest = "report-request"
	DesireRequest = "desire-request"
)

//go:generate mockgen -destination=../mock/sync.go -package=mock github.com/baetyl/baetyl/sync Sync
type Sync interface {
	Start()
	Close()
	Report(r v1.Report) (v1.Desire, error)
	SyncResource(v1.AppInfo) error
	SyncApps(infos []v1.AppInfo) (map[string]v1.Application, error)
}

// Sync sync shadow and resources with cloud
type sync struct {
	cfg    config.SyncConfig
	link   plugin.Link
	store  *bh.Store
	nod    *node.Node
	tomb   utils.Tomb
	log    *log.Logger
	http   *http.Client
	isSync bool
}

// NewSync create a new sync
func NewSync(cfg config.SyncConfig, store *bh.Store, nod *node.Node) (Sync, error) {
	link, err := plugin.GetPlugin("link")
	if err != nil {
		return nil, errors.Trace(err)
	}
	ops, err := cfg.Cloud.HTTP.ToClientOptions()
	if err != nil {
		return nil, errors.Trace(err)
	}
	s := &sync{
		cfg:   cfg,
		http:  http.NewClient(ops),
		store: store,
		nod:   nod,
		link:  link.(plugin.Link),
		log:   log.With(log.Any("core", "sync")),
	}
	s.isSync = s.link.IsAsyncSupported()
	return s, nil
}

func (s *sync) Start() {
	s.tomb.Go(s.reporting)
}

func (s *sync) Close() {
	s.tomb.Kill(nil)
	s.tomb.Wait()
}

func (s *sync) Report(r v1.Report) (v1.Desire, error) {
	msg := &plugin.Message{
		URI:     s.cfg.Cloud.Report.URL,
		Header:  map[string]string{RequestType: ReportRequest},
		Content: r,
	}
	res, err := s.link.Request(msg)
	if err != nil {
		return nil, errors.Trace(err)
	}
	s.log.Debug("sync reports cloud shadow", log.Any("report", msg))
	desire, ok := res.Content.(v1.Desire)
	if !ok {
		return nil, fmt.Errorf("unrecognized desire data")
	}
	return desire, nil
}

func (s *sync) reporting() error {
	s.log.Info("sync starts to report")
	defer s.log.Info("sync has stopped reporting")

	err := s.reportAndDesireAsync()
	if err != nil {
		s.log.Error("failed to report cloud shadow", log.Error(err))
	}

	t := time.NewTicker(s.cfg.Cloud.Report.Interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
			err := s.reportAndDesireAsync()
			if err != nil {
				s.log.Error("failed to report cloud shadow", log.Error(err))
			}
		case <-s.tomb.Dying():
			return nil
		}
	}
}

func (s *sync) reportAndDesireAsync() error {
	sd, err := s.nod.Get()
	if err != nil {
		return errors.Trace(err)
	}
	desire, err := s.Report(sd.Report)
	if err != nil {
		return errors.Trace(err)
	}
	if len(desire) == 0 {
		return nil
	}
	_, err = s.nod.Desire(desire)
	if err != nil {
		s.log.Error("failed to persist shadow desire", log.Any("desire", desire), log.Error(err))
		return errors.Trace(err)
	}
	return nil
}
