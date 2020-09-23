package sync

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/baetyl/baetyl/config"
	"github.com/baetyl/baetyl/mock/helper"
	"github.com/baetyl/baetyl/mock/plugin"
	"github.com/baetyl/baetyl/node"
	"github.com/baetyl/baetyl/store"
)

func TestReportSync(t *testing.T) {
	f, err := ioutil.TempFile("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())

	sto, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, sto)

	nod, err := node.NewNode(sto)
	assert.NoError(t, err)
	assert.NotNil(t, nod)

	var sc config.SyncConfig
	err = utils.UnmarshalYAML(nil, &sc)
	assert.NoError(t, err)

	mockCtl := gomock.NewController(t)
	link := plugin.NewMockLink(mockCtl)
	assert.NoError(t, err)
	syn := &sync{
		link:  link,
		cfg:   sc,
		store: sto,
		nod:   nod,
		log:   log.With(log.Any("test", "sync")),
	}
	link.EXPECT().IsAsyncSupported().Return(false)
	desire := specv1.Desire{"apps": map[string]interface{}{"app1": "123"}}

	msg := &specv1.Message{Content: specv1.LazyValue{Value: desire}, Kind: specv1.MessageReport}
	dt, err := json.Marshal(msg)
	assert.NoError(t, err)
	m := &specv1.Message{}
	err = json.Unmarshal(dt, m)
	assert.NoError(t, err)

	link.EXPECT().Request(gomock.Any()).Return(m, nil)
	err = syn.reportAndDesire()
	assert.NoError(t, err)
	no, _ := syn.nod.Get()
	assert.Equal(t, desire, no.Desire)

	link.EXPECT().IsAsyncSupported().Return(true)
	link.EXPECT().Send(gomock.Any()).Return(nil)
	err = syn.reportAndDesire()
	assert.NoError(t, err)

	link.EXPECT().IsAsyncSupported().Return(false)
	link.EXPECT().Request(gomock.Any()).Return(nil, errors.New("failed to report"))
	err = syn.reportAndDesire()
	assert.Error(t, err)

	link.EXPECT().IsAsyncSupported().Return(true)
	link.EXPECT().Send(gomock.Any()).Return(errors.New("failed to report"))
	assert.Error(t, err)
}

func TestReportAsync(t *testing.T) {
	f, err := ioutil.TempFile("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())

	sto, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, sto)

	nod, err := node.NewNode(sto)
	assert.NoError(t, err)
	assert.NotNil(t, nod)

	var sc config.SyncConfig
	err = utils.UnmarshalYAML(nil, &sc)
	assert.NoError(t, err)

	mockCtl := gomock.NewController(t)
	link := plugin.NewMockLink(mockCtl)

	hp := helper.NewMockHelper(mockCtl)
	hp.EXPECT().Subscribe("upside", gomock.Any()).Times(1)
	hp.EXPECT().Unsubscribe("upside").Times(1)

	syn := &sync{
		link:  link,
		cfg:   sc,
		store: sto,
		nod:   nod,
		hp:    hp,
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
