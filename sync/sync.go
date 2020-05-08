package sync

import (
	"crypto/x509"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/node"
	"github.com/baetyl/baetyl-go/http"
	"github.com/baetyl/baetyl-go/log"
	v1 "github.com/baetyl/baetyl-go/spec/v1"
	"github.com/baetyl/baetyl-go/utils"
	bh "github.com/timshannon/bolthold"
	"k8s.io/apimachinery/pkg/util/rand"
)

// ErrSyncTLSConfigMissing certificate bidirectional authentication is required for connection with cloud
var ErrSyncTLSConfigMissing = errors.New("certificate bidirectional authentication is required for connection with cloud")

const EnvKeyNodeName = "BAETYL_NODE_NAME"

// Sync sync shadow and resources with cloud
type Sync struct {
	cfg   config.SyncConfig
	fifo  chan v1.Desire
	http  *http.Client
	store *bh.Store
	nod   *node.Node
	tomb  utils.Tomb
	log   *log.Logger
}

// NewSync create a new sync
func NewSync(cfg config.SyncConfig, store *bh.Store, nod *node.Node) (*Sync, error) {
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
		nod:   nod,
		http:  http.NewClient(ops),
		fifo:  make(chan v1.Desire, 1),
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
			}
		} else {
			s.log.Error("certificate format error")
		}
	}
	return s, nil
}

func (s *Sync) Start() {
	s.tomb.Go(s.reporting, s.desiring)
}

func (s *Sync) Close() {
	s.tomb.Kill(nil)
	s.tomb.Wait()
}

func (s *Sync) ReportAndDesire() error {
	desire, err := s.report()
	if err != nil {
		return err
	}
	if len(desire) == 0 {
		return nil
	}

	err = s.syncResources(desire.AppInfos())
	if err != nil {
		return err
	}
	err = s.syncResources(desire.SysAppInfos())
	if err != nil {
		return err
	}
	_, err = s.nod.Desire(desire)
	if err != nil {
		return err
	}

	return nil
}

func (s *Sync) reporting() error {
	s.log.Info("sync starts to report")
	defer s.log.Info("sync has stopped reporting")

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

func (s *Sync) reportAndDesireAsync() error {
	desire, err := s.report()
	if err != nil {
		return err
	}
	if len(desire) == 0 {
		return nil
	}

	select {
	case s.fifo <- desire:
	case e := <-s.fifo:
		s.log.Info("ignore shadow desire", log.Any("desire", e))
		s.fifo <- desire
	case <-s.tomb.Dying():
	}
	return nil
}

func (s *Sync) report() (v1.Desire, error) {
	sd, err := s.nod.Get()
	if err != nil {
		return nil, err
	}
	pld, err := json.Marshal(sd.Report)
	if err != nil {
		return nil, err
	}
	s.log.Debug("sync reports cloud shadow", log.Any("report", sd.Report))
	data, err := s.http.PostJSON(s.cfg.Cloud.Report.URL, pld)
	if err != nil {
		return nil, err
	}
	var desire v1.Desire
	err = json.Unmarshal(data, &desire)
	if err != nil {
		return nil, err
	}
	return desire, nil
}

func (s *Sync) desiring() error {
	s.log.Info("sync starts to desire")
	defer s.log.Info("sync has stopped desiring")

	for {
		select {
		case e := <-s.fifo:
			// to prepare resources
			err := s.syncResources(e.AppInfos())
			if err != nil {
				s.log.Error("failed to sync resources", log.Any("desire", e), log.Error(err))
				continue
			}
			// to prepare resources
			err = s.syncResources(e.SysAppInfos())
			if err != nil {
				s.log.Error("failed to sync sys resources", log.Any("desire", e), log.Error(err))
				continue
			}
			// to persist desire
			_, err = s.nod.Desire(e)
			if err != nil {
				s.log.Error("failed to persist shadow desire", log.Any("desire", e), log.Error(err))
				continue
			}
		case <-s.tomb.Dying():
			return nil
		}
	}
}
