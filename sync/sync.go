package sync

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/http"
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
	cfg   config.SyncConfig
	http  *http.Client
	store *bh.Store
	nod   *node.Node
	tomb  utils.Tomb
	log   *log.Logger
}

// NewSync create a new sync
func NewSync(cfg config.SyncConfig, store *bh.Store, nod *node.Node) (Sync, error) {
	ops, err := cfg.Cloud.HTTP.ToClientOptions()
	if err != nil {
		return nil, errors.Trace(err)
	}
	if ops.TLSConfig == nil {
		return nil, errors.Trace(ErrSyncTLSConfigMissing)
	}
	s := &sync{
		cfg:   cfg,
		store: store,
		nod:   nod,
		http:  http.NewClient(ops),
		log:   log.With(log.Any("core", "sync")),
	}
	if len(ops.TLSConfig.Certificates) == 1 && len(ops.TLSConfig.Certificates[0].Certificate) == 1 {
		cert, err := x509.ParseCertificate(ops.TLSConfig.Certificates[0].Certificate[0])
		if err == nil {
			res := strings.SplitN(cert.Subject.CommonName, ".", 2)
			if len(res) != 2 || res[0] == "" || res[1] == "" {
				s.log.Error("failed to parse node name from cert")
			} else {
				os.Setenv(EnvKeyNodeName, res[1])
				os.Setenv(EnvKeyNodeNamespace, res[0])
			}
		} else {
			s.log.Error("certificate format error")
		}
	}
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
	pld, err := json.Marshal(r)
	if err != nil {
		return nil, errors.Trace(err)
	}
	s.log.Debug("sync reports cloud shadow", log.Any("report", string(pld)))
	data, err := s.http.PostJSON(s.cfg.Cloud.Report.URL, pld)
	if err != nil {
		return nil, errors.Trace(err)
	}
	var desire v1.Desire
	err = json.Unmarshal(data, &desire)
	if err != nil {
		return nil, errors.Trace(err)
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
