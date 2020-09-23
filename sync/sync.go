package sync

import (
	"time"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/log"
	v2plugin "github.com/baetyl/baetyl-go/v2/plugin"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
	bh "github.com/timshannon/bolthold"
	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/baetyl/baetyl/config"
	"github.com/baetyl/baetyl/helper"
	"github.com/baetyl/baetyl/node"
	"github.com/baetyl/baetyl/plugin"
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
	link  plugin.Link
	store *bh.Store
	nod   *node.Node
	tomb  utils.Tomb
	log   *log.Logger
	// for downloading objects
	download *http.Client
	hp       helper.Helper
}

// NewSync create a new sync
func NewSync(cfg config.Config, store *bh.Store, nod *node.Node, helper helper.Helper) (Sync, error) {
	link, err := v2plugin.GetPlugin(cfg.Plugin.Link)
	if err != nil {
		return nil, errors.Trace(err)
	}
	ops, err := cfg.Sync.Download.ToClientOptions()
	if err != nil {
		return nil, errors.Trace(err)
	}
	s := &sync{
		cfg:      cfg.Sync,
		download: http.NewClient(ops),
		store:    store,
		nod:      nod,
		link:     link.(plugin.Link),
		hp:       helper,
		log:      log.With(log.Any("core", "sync")),
	}
	return s, nil
}

func (s *sync) Start() {
	if s.link.IsAsyncSupported() {
		s.tomb.Go(s.receiving)
	}
	s.tomb.Go(s.reporting)
}

func (s *sync) receiving() error {
	s.hp.Subscribe(helper.TopicUpside, &handler{link: s.link})
	defer s.hp.Unsubscribe(helper.TopicUpside)

	msgCh, errCh := s.link.Receive()
	for {
		select {
		case <-s.tomb.Dying():
			return nil
		case msg := <-msgCh:
			err := s.dispatch(msg)
			if err != nil {
				s.log.Error("failed to dispatch message", log.Error(err))
			}
			continue
		case err := <-errCh:
			if err != nil {
				s.log.Error("failed to receive msg", log.Error(err))
				continue
			}
		}
	}
}

func (s *sync) dispatch(msg *v1.Message) error {
	switch msg.Kind {
	case v1.MessageReport:
		desire := v1.Desire{}
		err := msg.Content.Unmarshal(&desire)
		if err != nil {
			s.log.Error("receive unrecognized desire data", log.Error(err))
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
	case v1.MessageCMD:
		msg.Content.Value = map[string]string{}
		return s.publish(msg)
	case v1.MessageData:
		msg.Content.Value = []byte{}
		return s.publish(msg)
	default:
	}
	return nil
}

func (s *sync) publish(msg *v1.Message) error {
	err := msg.Content.Unmarshal(&msg.Content.Value)
	if err != nil {
		s.log.Error("failed to unmarshal message", log.Error(err))
		return errors.Trace(err)
	}
	err = s.hp.Publish(helper.TopicDownside, msg)
	if err != nil {
		s.log.Error("failed to publish message", log.Error(err))
		return errors.Trace(err)
	}
	return nil
}

func (s *sync) Close() {
	s.tomb.Kill(nil)
	s.tomb.Wait()
}

func (s *sync) reportAsync(r v1.Report) error {
	msg := &v1.Message{
		Kind:    v1.MessageReport,
		Content: v1.LazyValue{Value: r},
	}
	err := s.link.Send(msg)
	if err != nil {
		return errors.Trace(err)
	}
	s.log.Debug("reports cloud shadow async", log.Any("report", msg))
	return nil
}

func (s *sync) Report(r v1.Report) (v1.Desire, error) {
	msg := &v1.Message{
		Kind:    v1.MessageReport,
		Content: v1.LazyValue{Value: r},
	}
	res, err := s.link.Request(msg)
	if err != nil {
		return nil, errors.Trace(err)
	}
	s.log.Debug("sync reports cloud shadow", log.Any("report", msg))
	desire := v1.Desire{}
	err = res.Content.Unmarshal(&desire)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return desire, nil
}

func (s *sync) reporting() error {
	s.log.Info("sync starts to report")
	defer s.log.Info("sync has stopped reporting")

	err := s.reportAndDesire()
	if err != nil {
		s.log.Error("failed to report cloud shadow", log.Error(err))
	}

	t := time.NewTicker(s.cfg.Report.Interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
			err := s.reportAndDesire()
			if err != nil {
				s.log.Error("failed to report cloud shadow", log.Error(err))
			}
		case <-s.tomb.Dying():
			return nil
		}
	}
}

func (s *sync) reportAndDesire() error {
	sd, err := s.nod.Get()
	if err != nil {
		return errors.Trace(err)
	}
	if s.link.IsAsyncSupported() {
		err := s.reportAsync(sd.Report)
		if err != nil {
			return errors.Trace(err)
		}
	} else {
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
	}
	return nil
}
