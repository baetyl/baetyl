package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/event"
	"github.com/baetyl/baetyl-core/mock"
	"github.com/baetyl/baetyl-core/shadow"
	"github.com/baetyl/baetyl-core/store"
	v1 "github.com/baetyl/baetyl-go/spec/v1"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func prepare(t *testing.T) (*shadow.Shadow, *event.Center, config.EngineConfig) {
	f, err := ioutil.TempFile("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())

	sto, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, sto)

	sha, err := shadow.NewShadow(t.Name(), t.Name(), sto)
	assert.NoError(t, err)
	assert.NotNil(t, sha)

	cent, err := event.NewCenter(sto, 2)
	assert.NoError(t, err)
	assert.NotNil(t, cent)

	cfg := config.EngineConfig{
		Kind: "kubernetes",
	}
	cfg.Collector.Interval = time.Second
	return sha, cent, cfg
}

func TestReport(t *testing.T) {
	sha, cent, cfg := prepare(t)
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockAmi := mock.NewMockAMI(mockCtl)
	r := v1.Report{
		"apps": []interface{}{
			map[string]string{
				"name":    "app",
				"version": "v1",
			},
		},
	}
	mockAmi.EXPECT().Collect().Return(r, nil)

	wrongCfg := cfg
	wrongCfg.Kind = ""
	_, err := NewEngine(wrongCfg, mockAmi, sha, cent)
	assert.Error(t, err, os.ErrInvalid.Error())

	engine, err := NewEngine(cfg, mockAmi, sha, cent)
	assert.NoError(t, err)

	_, err = sha.Desire(v1.Desire{
		"apps": []interface{}{
			map[string]string{
				"name":    "app",
				"version": "v1",
			},
		},
	})
	assert.NoError(t, err)
	syn := mockSync{msgs: make(chan *event.Event, 1)}
	err = engine.cent.Register(event.SyncReportEvent, syn.report)
	cent.Start()
	assert.NoError(t, err)
	pld, _ := json.Marshal(r)
	expected := event.NewEvent(event.SyncReportEvent, pld)
	var msg *event.Event
	select {
	case msg = <-syn.msgs:
	}
	assert.Equal(t, msg.Payload, expected.Payload)
	engine.Close()
}

type mockSync struct {
	msgs chan *event.Event
}

func (s *mockSync) report(msg *event.Event) error {
	s.msgs <- msg
	return nil
}

func (s *mockSync) desire(msg *event.Event) error {
	s.msgs <- msg
	return nil
}

func TestDesire(t *testing.T) {
	sha, cent, cfg := prepare(t)
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockAmi := mock.NewMockAMI(mockCtl)
	r := v1.Report{
		"apps": []interface{}{
			map[string]string{
				"name":    "app",
				"version": "v1",
			},
		},
	}
	mockAmi.EXPECT().Collect().Return(r, nil)

	engine, err := NewEngine(cfg, mockAmi, sha, cent)
	assert.NoError(t, err)

	d := v1.Desire{
		"apps": []interface{}{
			map[string]string{
				"name":    "app",
				"version": "v2",
			},
		},
	}
	_, err = sha.Desire(d)
	assert.NoError(t, err)
	syn := mockSync{msgs: make(chan *event.Event, 1)}
	err = engine.cent.Register(event.SyncDesireEvent, syn.desire)
	cent.Start()
	assert.NoError(t, err)
	pld, _ := json.Marshal(d)
	expected := event.NewEvent(event.SyncReportEvent, pld)
	var msg *event.Event
	select {
	case msg = <-syn.msgs:
	}
	assert.Equal(t, msg.Payload, expected.Payload)
	engine.Close()
}

func TestApply(t *testing.T) {
	sha, cent, cfg := prepare(t)
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockAmi := mock.NewMockAMI(mockCtl)
	mockAmi.EXPECT().Apply(gomock.Any()).Return(nil)
	engine, err := NewEngine(cfg, mockAmi, sha, cent)
	assert.NoError(t, err)

	e := event.NewEvent("", []byte{})
	err = engine.Apply(e)
	assert.Error(t, err)

	wrongInfo := v1.Desire{
		"apps": []interface{}{},
	}
	pld, _ := json.Marshal(wrongInfo)
	e = event.NewEvent("", pld)
	err = engine.Apply(e)
	assert.Error(t, err)

	info := v1.Desire{
		"apps": []interface{}{
			map[string]string{
				"name":    "app",
				"version": "v1",
			},
		},
	}
	pld, _ = json.Marshal(info)
	e = event.NewEvent("", pld)
	err = engine.Apply(e)
	assert.NoError(t, err)
	engine.Close()

	err1 := errors.New("failed to apply")
	mockAmi.EXPECT().Apply(gomock.Any()).Return(err1)
	pld, _ = json.Marshal(info)
	e = event.NewEvent("", pld)
	err = engine.Apply(e)
	assert.Error(t, err, err1.Error())
	engine.Close()

}
