package initz

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
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
	"github.com/baetyl/baetyl/v2/mock"
	"github.com/baetyl/baetyl/v2/node"
	"github.com/baetyl/baetyl/v2/store"
)

func TestNewInitialize(t *testing.T) {
	cfg := config.Config{}
	// with cert
	tmpDir, err := ioutil.TempDir("", "init")
	assert.Nil(t, err)
	defer os.RemoveAll(tmpDir)

	certPath := path.Join(tmpDir, "cert.pem")
	err = ioutil.WriteFile(certPath, []byte("cert"), 0755)
	assert.Nil(t, err)
	cfg.Node.Cert = certPath

	f, err := ioutil.TempFile("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())
	cfg.Store.Path = f.Name()

	// bad case sync no plugin
	_, err = NewInitialize(cfg)
	assert.Error(t, err)

	// bad case engine ami err
	cfg.Plugin.Link = "link"
	v2plugin.RegisterFactory("link", func() (v2plugin.Plugin, error) {
		return &mockLink{}, nil
	})
	cfg.Plugin.Pubsub = "pubsub"
	v2plugin.RegisterFactory("pubsub", func() (v2plugin.Plugin, error) {
		res, err := pubsub.NewPubsub(1)
		assert.NoError(t, err)
		return res, nil
	})

	_, err = NewInitialize(cfg)
	assert.Error(t, err)

	// bad case no cert
	cfg.Node.Cert = ""

	_, err = NewInitialize(cfg)
	assert.Error(t, err)
}

func TestInitialize_start(t *testing.T) {
	ctl := gomock.NewController(t)
	eng := mock.NewMockEngine(ctl)
	syn := mock.NewMockSync(ctl)

	f, err := ioutil.TempFile("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())

	sto, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, sto)

	sha, err := node.NewNode(sto, nil)
	assert.NoError(t, err)
	assert.NotNil(t, sha)

	cfg := config.Config{}
	cfg.Sync.Report.Interval = time.Nanosecond

	init := &Initialize{
		cfg:  cfg,
		sto:  sto,
		sha:  sha,
		eng:  eng,
		syn:  syn,
		log:  log.L().With(),
		tomb: utils.Tomb{},
	}

	ds := specv1.Desire{}
	ds.SetAppInfos(true, []specv1.AppInfo{{
		Name:    "baetyl-core",
		Version: "123",
	}})
	r := specv1.Report{}
	syn.EXPECT().Report(r).Return(specv1.Delta(ds), nil).Times(1)
	syn.EXPECT().Report(r).Return(nil, os.ErrInvalid).Times(1)
	eng.EXPECT().ReportAndDesire().Return(nil).Times(1)
	eng.EXPECT().Collect("baetyl-edge-system", true, nil).Return(specv1.Report{}).Times(2)

	err = init.start()
	assert.Error(t, err, os.ErrInvalid)
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
