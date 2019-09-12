package main

import (
	"fmt"

	"github.com/baetyl/baetyl/sdk/baetyl-go"
	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/cfgfile"
	"github.com/elastic/beats/libbeat/cmd/instance"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/logp/configure"
	"github.com/elastic/beats/libbeat/paths"
	"github.com/elastic/beats/libbeat/publisher/pipeline"
	_ "github.com/elastic/beats/libbeat/publisher/queue/memqueue"
	"github.com/elastic/beats/libbeat/version"
)

// The name of filebeat
const (
	FILEBEATNAME = "filebeat"
)

func newFilebeat() (*beat.Beat, error) {
	b, err := instance.NewBeat(FILEBEATNAME, FILEBEATNAME, version.GetDefaultVersion())
	if err != nil {
		panic(err)
	}

	cfg, err := cfgfile.Load(baetyl.DefaultFilebeatConfFile)
	b.RawConfig = cfg
	err = cfg.Unpack(&b.Config)
	if err != nil {
		return nil, fmt.Errorf("error loading config file: %v", err)
	}

	b.Beat.Config = &b.Config.BeatConfig

	err = paths.InitPaths(&b.Config.Path)
	if err != nil {
		return nil, fmt.Errorf("error setting default paths: %v", err)
	}

	var Logging *common.Config
	if err := configure.Logging(b.Beat.Info.Beat, Logging); err != nil {
		return nil, fmt.Errorf("error initializing logging: %v", err)
	}

	logp.Info(paths.Paths.String())

	b.Beat.BeatConfig, err = b.BeatConfig()
	if err != nil {
		return nil, err
	}

	pipeline, err := pipeline.Load(b.Beat.Info, nil, b.Config.Pipeline, b.Config.Output)
	if err != nil {
		return nil, fmt.Errorf("error initializing publisher: %v", err)
	}

	b.Beat.Publisher = pipeline
	return &b.Beat, nil
}

func (a *agent) filebeting() error {
	a.ctx.Log().Infof("%s start running.", a.beat.Info.Beat)
	err := a.beater.Run(a.beat)
	if err != nil {
		a.ctx.Log().Errorf("failed to start filebeat", err)
	}
	return nil
}
