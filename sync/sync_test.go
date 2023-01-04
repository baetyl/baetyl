package sync

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/baetyl/baetyl-go/v2/log"
	v2plugin "github.com/baetyl/baetyl-go/v2/plugin"
	"github.com/baetyl/baetyl-go/v2/pubsub"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/baetyl/baetyl/v2/config"
	"github.com/baetyl/baetyl/v2/mock/plugin"
	"github.com/baetyl/baetyl/v2/node"
	"github.com/baetyl/baetyl/v2/store"
)

func TestReportSync(t *testing.T) {
	f, err := os.CreateTemp("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())

	sto, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, sto)

	nod, err := node.NewNode(sto)
	assert.NoError(t, err)
	assert.NotNil(t, nod)

	var sc config.Config
	err = utils.UnmarshalYAML(nil, &sc)
	assert.NoError(t, err)

	mockCtl := gomock.NewController(t)
	link := plugin.NewMockLink(mockCtl)
	assert.NoError(t, err)
	pb := plugin.NewMockPubsub(mockCtl)
	syn := &sync{
		link:  link,
		cfg:   sc,
		store: sto,
		nod:   nod,
		pb:    pb,
		log:   log.With(log.Any("test", "sync")),
	}
	link.EXPECT().IsAsyncSupported().Return(false).Times(1)
	delta := specv1.Delta{"apps": map[string]interface{}{"app1": "123"}}

	msg := &specv1.Message{Content: specv1.LazyValue{Value: delta}, Kind: specv1.MessageReport}
	dt, err := json.Marshal(msg)
	assert.NoError(t, err)
	m := &specv1.Message{}
	err = json.Unmarshal(dt, m)
	assert.NoError(t, err)

	link.EXPECT().Request(gomock.Any()).Return(m, nil).Times(1)
	err = syn.reportAndDesire()
	assert.NoError(t, err)
	no, _ := syn.nod.Get()
	var desire specv1.Desire = map[string]interface{}{}
	desire, err = desire.Patch(delta)
	assert.NoError(t, err)
	assert.Equal(t, desire, no.Desire)

	link.EXPECT().IsAsyncSupported().Return(true).Times(1)
	link.EXPECT().Send(gomock.Any()).Return(nil).Times(1)
	err = syn.reportAndDesire()
	assert.NoError(t, err)

	link.EXPECT().IsAsyncSupported().Return(false).Times(1)
	link.EXPECT().Request(gomock.Any()).Return(nil, errors.New("failed to report"))
	err = syn.reportAndDesire()
	assert.Error(t, err)

	link.EXPECT().IsAsyncSupported().Return(true).Times(1)
	link.EXPECT().Send(gomock.Any()).Return(errors.New("failed to report"))
	err = syn.reportAndDesire()
	assert.Error(t, err)
}

func TestReportAsync(t *testing.T) {
	f, err := os.CreateTemp("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())

	sto, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, sto)

	nod, err := node.NewNode(sto)
	assert.NoError(t, err)
	assert.NotNil(t, nod)

	var sc config.Config
	err = utils.UnmarshalYAML(nil, &sc)
	assert.NoError(t, err)

	mockCtl := gomock.NewController(t)
	link := plugin.NewMockLink(mockCtl)
	pb := plugin.NewMockPubsub(mockCtl)

	ch := make(<-chan interface{})
	pb.EXPECT().Subscribe("upside").Return(ch, nil).Times(1)
	pb.EXPECT().Unsubscribe("upside", ch).Return(nil).Times(1)

	syn := &sync{
		link:  link,
		cfg:   sc,
		store: sto,
		nod:   nod,
		pb:    pb,
		log:   log.With(log.Any("test", "sync")),
	}
	desire := specv1.Desire{"apps": map[string]interface{}{"app1": "123"}}
	bt, err := json.Marshal(desire)
	assert.NoError(t, err)

	m := &specv1.Message{Content: specv1.LazyValue{}, Kind: specv1.MessageReport}
	dt, err := json.Marshal(m)
	assert.NoError(t, err)
	msg := &specv1.Message{}
	err = json.Unmarshal(dt, msg)
	assert.NoError(t, err)

	err = msg.Content.UnmarshalJSON(bt)
	assert.NoError(t, err)

	msgCh := make(chan *specv1.Message, 1)
	msgCh <- msg
	errCh := make(chan error, 1)
	errCh <- nil
	link.EXPECT().IsAsyncSupported().Return(true).Times(2)
	link.EXPECT().Send(gomock.Any()).Return(nil)
	link.EXPECT().Receive().Return(msgCh, errCh)
	syn.Start()
	time.Sleep(time.Millisecond * 500)
	no, _ := syn.nod.Get()
	assert.Equal(t, desire, no.Desire)
	syn.Close()
}

func TestSync_dispatch(t *testing.T) {
	pb, err := pubsub.NewPubsub(1)
	assert.NoError(t, err)

	s := &sync{
		log: log.L().With(),
		pb:  pb,
	}

	msg := &specv1.Message{
		Kind: specv1.MessageCMD,
	}

	err = s.dispatch(msg)
	assert.NoError(t, err)
}

func TestNewSync(t *testing.T) {
	cfg := config.Config{}

	// bad case 0
	_, err := NewSync(cfg, nil, nil)
	assert.Error(t, err)

	// bad case 1
	cfg.Plugin.Link = "link"
	v2plugin.RegisterFactory("link", func() (v2plugin.Plugin, error) {
		return &mockLink{}, nil
	})
	_, err = NewSync(cfg, nil, nil)
	assert.Error(t, err)

	// good case
	cfg.Plugin.Pubsub = "pubsub"
	v2plugin.RegisterFactory("pubsub", func() (v2plugin.Plugin, error) {
		res, err := pubsub.NewPubsub(1)
		assert.NoError(t, err)
		return res, nil
	})

	_, err = NewSync(cfg, nil, nil)
	assert.NoError(t, err)

	// bad case 2
	cfg.Sync.Download.Cert = "NotExist"
	cfg.Sync.Download.CA = "NotExist"
	cfg.Sync.Download.Key = "NotExist"

	_, err = NewSync(cfg, nil, nil)
	assert.Error(t, err)
}

type mockLink struct{}

func (lk *mockLink) Close() error {
	return nil
}

func (lk *mockLink) Receive() (<-chan *specv1.Message, <-chan error) {
	return nil, nil
}

func (lk *mockLink) Request(msg *specv1.Message) (*specv1.Message, error) {
	return nil, nil
}

func (lk *mockLink) Send(msg *specv1.Message) error {
	return nil
}

func (lk *mockLink) IsAsyncSupported() bool {
	return false
}

func (lk *mockLink) State() *specv1.Message {
	return nil
}
