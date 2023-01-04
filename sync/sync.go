package sync

import (
	"os"
	"time"

	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/log"
	goplugin "github.com/baetyl/baetyl-go/v2/plugin"
	"github.com/baetyl/baetyl-go/v2/pubsub"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
	bh "github.com/timshannon/bolthold"
	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/baetyl/baetyl/v2/config"
	"github.com/baetyl/baetyl/v2/eventx"
	"github.com/baetyl/baetyl/v2/node"
	"github.com/baetyl/baetyl/v2/plugin"
)

const (
	EnvKeyNodeName      = "BAETYL_NODE_NAME"
	EnvKeyNodeNamespace = "BAETYL_NODE_NAMESPACE"

	TopicUpside   = "upside"
	TopicDownside = "downside"

	TopicDM = "dm"

	BaetylHookUploadObject = "baetyl_upload_object"
)

var Hooks = map[string]interface{}{}

type UploadObjectFunc func(dir, file, md5, unpack string) error

//go:generate mockgen -destination=../mock/sync.go -package=mock -source=sync.go Sync
type Sync interface {
	Start()
	Close()
	Report(r v1.Report) (v1.Desire, error)
	SyncResource(v1.AppInfo) error
	SyncApps(infos []v1.AppInfo) (map[string]v1.Application, error)
}

// Sync sync shadow and resources with cloud
type sync struct {
	cfg   config.Config
	link  plugin.Link
	store *bh.Store
	nod   node.Node
	tomb  utils.Tomb
	log   *log.Logger
	// for downloading objects
	download *http.Client
	pb       plugin.Pubsub
}

// NewSync create a new sync
func NewSync(cfg config.Config, store *bh.Store, nod node.Node) (Sync, error) {
	link, err := goplugin.GetPlugin(cfg.Plugin.Link)
	if err != nil {
		return nil, errors.Trace(err)
	}
	pb, err := goplugin.GetPlugin(cfg.Plugin.Pubsub)
	if err != nil {
		return nil, errors.Trace(err)
	}
	ops, err := cfg.Sync.Download.ToClientOptions()
	if err != nil {
		return nil, errors.Trace(err)
	}
	s := &sync{
		cfg:      cfg,
		download: http.NewClient(ops),
		store:    store,
		nod:      nod,
		link:     link.(plugin.Link),
		pb:       pb.(plugin.Pubsub),
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
	s.log.Debug("subscribe upside")
	upsideChan, err := s.pb.Subscribe(TopicUpside)
	if err != nil {
		s.log.Error("failed to subscribe upside topic", log.Any("topic", TopicUpside), log.Error(err))
	}
	processor := pubsub.NewProcessor(upsideChan, 0, &handler{link: s.link})
	processor.Start()
	defer func() {
		s.log.Debug("unsubscribe upside")
		err = s.pb.Unsubscribe(TopicUpside, upsideChan)
		if err != nil {
			s.log.Error("failed to unsubscribe upside topic", log.Any("topic", TopicUpside), log.Error(err))
		}
		processor.Close()
	}()

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
		_, err := s.nod.Get()
		if err != nil {
			return errors.Trace(err)
		}
		desire := v1.Desire{}
		err = msg.Content.Unmarshal(&desire)
		if err != nil {
			s.log.Error("receive unrecognized desire data", log.Error(err))
			return errors.Trace(err)
		}
		if len(desire) == 0 {
			return nil
		}
		delta, err := s.nod.Desire(desire, true)
		if err != nil {
			s.log.Error("failed to persist shadow desire", log.Any("desire", desire), log.Error(err))
			return errors.Trace(err)
		}
		if v1.BaetylCore != os.Getenv(context.KeySvcName) {
			return nil
		}
		if err := s.sendEvent(delta); err != nil {
			return errors.Trace(err)
		}
		if devices, ok := delta[v1.KeyDevices].([]interface{}); ok && len(devices) > 0 {
			msg := &v1.Message{Kind: v1.MessageDevices, Content: v1.LazyValue{Value: devices}}
			s.log.Debug("sync devices msg", log.Any("devices msg", msg))
			return s.pb.Publish(TopicDM, msg)
		}
	case v1.MessageCMD, v1.MessageData:
		s.log.Debug("sync downside msg", log.Any("msg", msg))
		return s.pb.Publish(TopicDownside, msg)
	case v1.MessageDeviceDelta, v1.MessageDeviceEvent, v1.MessageDevicePropertyGet:
		s.log.Debug("sync dm msg", log.Any("msg", msg))
		return s.pb.Publish(TopicDM, msg)
	case v1.MessageNodeProps:
		// TODO only update node props
		desire := v1.Desire{}
		err := msg.Content.Unmarshal(&desire)
		if err != nil {
			return errors.Trace(err)
		}
		if len(desire) == 0 {
			return nil
		}
		delta, err := s.nod.Desire(desire, true)
		if err != nil {
			return errors.Trace(err)
		}
		if v1.BaetylCore != os.Getenv(context.KeySvcName) {
			return nil
		}
		if err := s.sendEvent(delta); err != nil {
			return errors.Trace(err)
		}
	default:
	}
	return nil
}

func (s *sync) sendEvent(delta v1.Delta) error {
	if s.cfg.Event.Notify {
		if props, ok := delta[v1.KeyNodeProps].(map[string]interface{}); ok && len(props) > 0 {
			msg := &v1.Message{Kind: v1.MessageNodeProps, Content: v1.LazyValue{Value: props}}
			s.log.Debug("sync node props", log.Any("node props msg", msg))
			return s.pb.Publish(eventx.TopicEvent, msg)
		}
	}
	return nil
}

func (s *sync) Close() {
	s.tomb.Kill(nil)
	s.tomb.Wait()
}

func (s *sync) reportAsync(r v1.Report) error {
	msg := &v1.Message{
		Kind:     v1.MessageReport,
		Metadata: map[string]string{"source": os.Getenv(context.KeySvcName)},
		Content:  v1.LazyValue{Value: r},
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
		Kind:     v1.MessageReport,
		Metadata: map[string]string{"source": os.Getenv(context.KeySvcName)},
		Content:  v1.LazyValue{Value: r},
	}
	res, err := s.link.Request(msg)
	if err != nil {
		return nil, errors.Trace(err)
	}
	s.log.Debug("sync reports cloud shadow", log.Any("report", msg))
	var desire v1.Desire
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

	t := time.NewTicker(s.cfg.Sync.Report.Interval)
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
	shadow, err := s.nod.Get()
	if err != nil {
		return errors.Trace(err)
	}
	if s.link.IsAsyncSupported() {
		reportErr := s.reportAsync(shadow.Report)
		if reportErr != nil {
			return errors.Trace(reportErr)
		}
	} else {
		desire, reportErr := s.Report(shadow.Report)
		if reportErr != nil {
			return errors.Trace(reportErr)
		}
		if len(desire) == 0 {
			return nil
		}
		delta, desErr := s.nod.Desire(desire, true)
		if desErr != nil {
			s.log.Error("failed to persist shadow desire", log.Any("desire", desire), log.Error(desErr))
			return errors.Trace(desErr)
		}
		if v1.BaetylCore != os.Getenv(context.KeySvcName) {
			return nil
		}
		if s.cfg.Event.Notify {
			if props, ok := delta[v1.KeyNodeProps].(map[string]interface{}); ok && len(props) > 0 {
				msg := &v1.Message{Kind: v1.MessageNodeProps, Content: v1.LazyValue{Value: props}}
				s.log.Debug("sync node props", log.Any("node props msg", msg))
				return s.pb.Publish(eventx.TopicEvent, msg)
			}
		}
		if devices, ok := delta[v1.KeyDevices].([]interface{}); ok && len(devices) > 0 {
			msg := &v1.Message{Kind: v1.MessageDevices, Content: v1.LazyValue{Value: devices}}
			s.log.Debug("sync devices msg", log.Any("devices msg", msg))
			return s.pb.Publish(TopicDM, msg)
		}
	}
	return nil
}
