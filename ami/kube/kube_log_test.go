package kube

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl/config"
	"github.com/baetyl/baetyl/store"
	"github.com/stretchr/testify/assert"
)

func TestToLogOptions(t *testing.T) {
	ami := initLogKubeAMI(t)
	opts := ami.toLogOptions(0, 0)
	fmt.Println(opts.TailLines)
	assert.Nil(t, opts.TailLines)
	assert.Nil(t, opts.SinceSeconds)
	assert.Equal(t, opts.Previous, ami.conf.LogConfig.Previous)
	assert.Equal(t, opts.Follow, ami.conf.LogConfig.Follow)
	assert.Equal(t, opts.Timestamps, ami.conf.LogConfig.TimeStamps)

	opts = ami.toLogOptions(int64(10), int64(60))
	fmt.Println(opts.TailLines)
	assert.Equal(t, *opts.TailLines, int64(10))
	assert.Equal(t, *opts.SinceSeconds, int64(60))
	assert.Equal(t, opts.Previous, ami.conf.LogConfig.Previous)
	assert.Equal(t, opts.Follow, ami.conf.LogConfig.Follow)
	assert.Equal(t, opts.Timestamps, ami.conf.LogConfig.TimeStamps)
}

func initLogKubeAMI(t *testing.T) *kubeImpl {
	f, err := ioutil.TempFile("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())
	sto, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, sto)
	return &kubeImpl{
		cli:   nil,
		store: sto,
		knn:   "node1",
		conf: &config.KubernetesConfig{
			LogConfig: config.KubernetesLogConfig{
				Follow:     false,
				Previous:   false,
				TimeStamps: false,
			},
		},
		log: log.With(),
	}
}
