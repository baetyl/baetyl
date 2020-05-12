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

var (
	// ErrSyncTLSConfigMissing certificate bidirectional authentication is required for connection with cloud
	ErrSyncTLSConfigMissing = errors.New("certificate bidirectional authentication is required for connection with cloud")
	// ErrSysappCoreMissing system application baetyl-core is required for connection with cloud
	ErrSysappCoreMissing = errors.New("system application baetyl-core is required for connection with cloud")
)

const (
	BaetylCore     = "baetyl-core"
	EnvKeyNodeName = "BAETYL_NODE_NAME"
)

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

func (s *Sync) Report(r v1.Report) error {
	ds, err := s.report(r)
	s.log.Debug("init report info", log.Any("Report", r))
	s.log.Debug("init desire info", log.Any("Report", ds))
	return err
}

func (s *Sync) ReportAndDesire() error {
	for {
		err := s.desireCore()
		if err != ErrSysappCoreMissing {
			return err
		}
		time.Sleep(s.cfg.Cloud.Report.Interval)
	}
}

func (s *Sync) desireCore() error {
	sd, err := s.nod.Get()
	if err != nil {
		return err
	}
	desire, err := s.report(sd.Report)
	if err != nil {
		s.log.Error("sync report error", log.Any("ReportAndDesire", err))
		return ErrSysappCoreMissing
	}
	if len(desire) == 0 || len(desire.SysAppInfos()) == 0 {
		return ErrSysappCoreMissing
	}

	for _, app := range desire.SysAppInfos() {
		if strings.Contains(app.Name, BaetylCore) {
			ds := v1.Desire{
				"sysapps": []v1.AppInfo{app},
			}
			err = s.syncResources(ds.SysAppInfos())
			if err != nil {
				return err
			}
			_, err = s.nod.Desire(ds)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return ErrSysappCoreMissing
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
	sd, err := s.nod.Get()
	if err != nil {
		return err
	}
	desire, err := s.report(sd.Report)
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

func (s *Sync) report(r v1.Report) (v1.Desire, error) {
	pld, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	s.log.Debug("sync reports cloud shadow", log.Any("report", string(pld)))
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
