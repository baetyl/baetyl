package sync

import (
	"errors"
	"fmt"
	"github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/baetyl/baetyl/config"
	"github.com/baetyl/baetyl/mock/plugin"
	"github.com/baetyl/baetyl/node"
	"github.com/baetyl/baetyl/store"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
	"time"
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
	msg := &specv1.Message{Content: desire, Kind: specv1.MessageReport}
	link.EXPECT().Request(gomock.Any()).Return(msg, nil)
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
	assert.NoError(t, err)
	syn := &sync{
		link:  link,
		cfg:   sc,
		store: sto,
		nod:   nod,
		log:   log.With(log.Any("test", "sync")),
	}
	desire := specv1.Desire{"apps": map[string]interface{}{"app1": "123"}}
	msg := &specv1.Message{Content: desire, Kind: specv1.MessageReport}
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
