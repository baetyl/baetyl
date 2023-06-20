package chain

import (
	"os"
	"sync"
	"testing"

	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/plugin"
	"github.com/baetyl/baetyl-go/v2/pubsub"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/baetyl/baetyl/v2/config"
	"github.com/baetyl/baetyl/v2/mock"
)

const (
	token   = "0123456789"
	testMsg = "test"
)

var (
	chainWG = sync.WaitGroup{}
)

func initChainEnv(t *testing.T) (config.Config, *gomock.Controller, map[string]string) {
	cfg := config.Config{}
	cfg.Plugin.Pubsub = "defaultpubsub"

	plugin.RegisterFactory("defaultpubsub", func() (plugin.Plugin, error) {
		res, err := pubsub.NewPubsub(1)
		assert.NoError(t, err)
		return res, nil
	})

	ctl := gomock.NewController(t)

	data := map[string]string{
		"namespace": "default",
		"name":      "baetyl-function-0",
		"container": "function",
		"token":     token,
	}

	return cfg, ctl, data
}

func TestNewChain(t *testing.T) {
	cfg, ctl, data := initChainEnv(t)
	ami := mock.NewMockAMI(ctl)

	t.Setenv(context.KeyRunMode, context.RunModeNative)
	c, err := NewChain(cfg, ami, data, true)
	assert.Error(t, err, ErrParseData)
	data["port"] = "22"
	c, err = NewChain(cfg, ami, data, true)
	assert.Error(t, err, ErrParseData)
	data["userName"] = "root"
	c, err = NewChain(cfg, ami, data, true)
	assert.Error(t, err, ErrParseData)
	data["password"] = "1234"
	c, err = NewChain(cfg, ami, data, true)
	assert.NoError(t, err)
	assert.NotNil(t, c)

	ami.EXPECT().RemoteCommand(gomock.Any(), gomock.Any()).Return(os.ErrInvalid).MaxTimes(1)
	err = c.Debug()
	assert.NoError(t, err)
	err = c.Close()
	assert.NoError(t, err)

	// websocket case
	data["host"] = "127.0.0.1"
	data["path"] = ""
	c, err = NewChain(cfg, ami, data, true)
	assert.NoError(t, err)

	// bad case
	cfg.Plugin.Pubsub = "not exist"
	_, err = NewChain(cfg, ami, data, true)
	assert.Error(t, err)
}

func TestChainMsg(t *testing.T) {
	cfg, ctl, data := initChainEnv(t)
	a := mock.NewMockAMI(ctl)

	c, err := NewChain(cfg, a, data, false)
	assert.NoError(t, err)
	assert.NotNil(t, c)

	cha, ok := c.(*chain)
	assert.True(t, ok)
	cha.upside = "chainUP"

	a.EXPECT().RemoteCommand(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	go func() {
		dt := make([]byte, 1024)
		n, err := cha.pipe.InReader.Read(dt)
		assert.NoError(t, err)
		assert.Equal(t, testMsg, string(dt[0:n]))
		_, err = cha.pipe.OutWriter.Write(dt[0:n])
		assert.NoError(t, err)
	}()

	ch, err := cha.pb.Subscribe(cha.upside)
	assert.NoError(t, err)
	assert.NotNil(t, ch)

	chainWG.Add(1)
	pro := pubsub.NewProcessor(ch, 0, &hpUpside{t: t, pb: cha.pb})
	pro.Start()

	chainWG.Add(1)
	err = c.Debug()
	assert.NoError(t, err)

	sendMsg := &specv1.Message{
		Kind:    specv1.MessageData,
		Content: specv1.LazyValue{Value: []byte(testMsg)},
	}
	err = cha.pb.Publish(cha.downside, sendMsg)
	assert.NoError(t, err)

	chainWG.Wait()
	err = c.Close()
	assert.NoError(t, err)
	pro.Close()
}

type hpUpside struct {
	t  *testing.T
	pb pubsub.Pubsub
}

func (h *hpUpside) OnMessage(msg interface{}) error {
	m, ok := msg.(*specv1.Message)
	assert.True(h.t, ok)
	switch m.Kind {
	case specv1.MessageCMD:
		assert.Equal(h.t, token, m.Metadata["token"])
		assert.Equal(h.t, "false", m.Metadata["success"])
		chainWG.Done()
	case specv1.MessageData:
		assert.Equal(h.t, token, m.Metadata["token"])
		var cmd []byte
		err := m.Content.Unmarshal(&cmd)
		assert.NoError(h.t, err)
		assert.Equal(h.t, testMsg, string(cmd))
		chainWG.Done()
	}
	return nil
}

func (h *hpUpside) OnTimeout() error {
	return nil
}
