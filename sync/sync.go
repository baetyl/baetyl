package sync

import (
	"encoding/json"
	"errors"
	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/node"
	"github.com/baetyl/baetyl-go/http"
	"github.com/baetyl/baetyl-go/log"
	v1 "github.com/baetyl/baetyl-go/spec/v1"
	"github.com/baetyl/baetyl-go/utils"
	bh "github.com/timshannon/bolthold"
	"k8s.io/apimachinery/pkg/util/rand"
	"time"
)

// ErrSyncTLSConfigMissing certificate bidirectional authentication is required for connection with cloud
var ErrSyncTLSConfigMissing = errors.New("Certificate bidirectional authentication is required for connection with cloud")

// Sync sync shadow and resources with cloud
type Sync struct {
	cfg   config.SyncConfig
	http  *http.Client
	store *bh.Store
	shad  *node.Node
	tomb  utils.Tomb
	log   *log.Logger
}

// NewSync create a new sync
func NewSync(cfg config.SyncConfig, store *bh.Store, shad *node.Node) (*Sync, error) {
	ops, err := cfg.Cloud.HTTP.ToClientOptions()
	if err != nil {
		return nil, err
	}
	if ops.TLSConfig == nil {
		return nil, ErrSyncTLSConfigMissing
	}
	s := &Sync{
		cfg:   cfg,
		store: store,
		shad:  shad,
		http:  http.NewClient(ops),
		log:   log.With(log.Any("core", "sync")),
	}
	s.tomb.Go(s.reporting)
	return s, nil
}

// Report reports info
func (s *Sync) reporting() error {
	s.log.Info("sync starts to report")
	defer s.log.Info("sync has stopped reporting")

	t := time.NewTicker(s.cfg.Cloud.Report.Interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
			err := s.report()
			if err != nil {
				s.log.Error("failed to report cloud shadow", log.Error(err))
			} else {
				s.log.Debug("sync reports cloud shadow")
			}
		case <-s.tomb.Dying():
			return nil
		}
	}
}

// Report reports info
func (s *Sync) report() error {
	sd, err := s.shad.Get()
	if err != nil {
		return err
	}
	pld, err := json.Marshal(sd.Report)
	if err != nil {
		return err
	}
	data, err := s.http.PostJSON(s.cfg.Cloud.Report.URL, pld)
	if err != nil {
		return err
	}
	var desire v1.Desire
	err = json.Unmarshal(data, &desire)
	if err != nil {
		return err
	}
	if len(desire) == 0 {
		return nil
	}

	// to prepare resources
	err = s.syncResources(desire.AppInfos())
	if err != nil {
		s.log.Error("failed to sync resources", log.Error(err))
		return err
	}

	// to persist desire
	_, err = s.shad.Desire(desire)
	if err != nil {
		s.log.Error("failed to persist shadow desire", log.Any("desire", desire), log.Error(err))
		return err
	}
	return nil
}
