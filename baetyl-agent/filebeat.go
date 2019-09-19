package main

import (
	"fmt"

	"github.com/elastic/beats/filebeat/beater"
	"github.com/elastic/beats/libbeat/beat"

	"github.com/baetyl/baetyl/sdk/baetyl-go"
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

type filebeat struct {
	//filebeat conf
	beat *beat.Beat
	//filebeat
	beater beat.Beater
}

func newFilebeat() (*filebeat, error) {
	beat, err := newBeat()
	if err != nil {
		return nil, err
	}
	beater, err := beater.New(beat, beat.BeatConfig)
	if err != nil {
		return nil, err
	}
	return &filebeat{
		beat:   beat,
		beater: beater,
	}, nil
}

func newBeat() (*beat.Beat, error) {
	b, err := instance.NewBeat(FILEBEATNAME, FILEBEATNAME, version.GetDefaultVersion())
	if err != nil {
		return nil, fmt.Errorf("failed to creates a new beat instance: %v", err)
	}

	b.RawConfig, err = cfgfile.Load(baetyl.DefaultFilebeatConfFile)
	err = b.RawConfig.Unpack(&b.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to loading config file: %v", err)
	}

	b.Beat.Config = &b.Config.BeatConfig

	err = paths.InitPaths(&b.Config.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to setting default paths: %v", err)
	}

	var Logging *common.Config
	if err := configure.Logging(b.Beat.Info.Beat, Logging); err != nil {
		return nil, fmt.Errorf("failed to initializing logging: %v", err)
	}

	logp.Info(paths.Paths.String())

	b.Beat.BeatConfig, err = b.BeatConfig()
	if err != nil {
		return nil, err
	}

	b.Beat.Publisher, err = pipeline.Load(b.Beat.Info, nil, b.Config.Pipeline, b.Config.Output)
	if err != nil {
		return nil, fmt.Errorf("failed to initializing publisher: %v", err)
	}

	return &b.Beat, nil
}

func (a *agent) filebeating() error {
	a.ctx.Log().Infof("%s start running, version is %s.", a.filebeat.beat.Info.Beat, a.filebeat.beat.Info.Version)
	err := a.filebeat.beater.Run(a.filebeat.beat)
	if err != nil {
		a.ctx.Log().Errorf("failed to start filebeat: %v", err)
	}
	return nil
}
