package logger

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestInitLogger(t *testing.T) {
	type args struct {
		vs []string
	}
	tests := []struct {
		name string
		args args
		want logrus.Fields
	}{
		{
			name: "0",
			args: args{
				vs: []string{},
			},
			want: logrus.Fields{},
		},
		{
			name: "1",
			args: args{
				vs: []string{"k1"},
			},
			want: logrus.Fields{},
		},
		{
			name: "2",
			args: args{
				vs: []string{"k1", "v2"},
			},
			want: logrus.Fields{"k1": "v2"},
		},
		{
			name: "3",
			args: args{
				vs: []string{"k1", "v2", "k3"},
			},
			want: logrus.Fields{"k1": "v2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := InitLogger(LogInfo{
				Level: "debug",
			}, tt.args.vs...)
			assert.EqualValues(t, l.(*logger).entry.Data, tt.want)
		})
	}
}

func TestLogger(t *testing.T) {
	logger := WithField("name", "baetyl")
	assert.NotEmpty(t, logger)

	logger = WithError(errors.New("test error"))
	assert.NotEmpty(t, logger)

	logger.Debugf("%s debug", "baetyl")
	logger.Infof("%s info", "baetyl")
	logger.Warnf("%s debug", "baetyl")
	logger.Errorf("%s errorf", "baetyl")

	logger.Debugln("baetyl debug")
	logger.Infoln("baetyl info")
	logger.Warnln("baetyl debug")
	logger.Errorln("baetyl errorf")

	Debugf("%s debug", "baetyl")
	Infof("%s info", "baetyl")
	Warnf("%s debug", "baetyl")
	Errorf("%s errorf", "baetyl")

	Debugln("baetyl debug")
	Infoln("baetyl info")
	Warnln("baetyl debug")
	Errorln("baetyl errorf")

	dir, err := ioutil.TempDir("", t.Name())
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	c := LogInfo{
		Path:   dir,
		Level:  "info",
		Format: "",
		Age: struct {
			Max int `yaml:"max" json:"max" default:"15" validate:"min=1"`
		}{
			Max: 1,
		},
		Size: struct {
			Max int `yaml:"max" json:"max" default:"50" validate:"min=1"`
		}{
			Max: 1,
		},
		Backup: struct {
			Max int `yaml:"max" json:"max" default:"15" validate:"min=1"`
		}{
			Max: 1,
		},
	}
	logger = New(c, "baetyl")
	assert.NotEmpty(t, logger)
}
