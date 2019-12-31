package main

import (
	"bufio"
	"encoding/json"
	"github.com/baetyl/baetyl/baetyl-agent/config"
	"github.com/baetyl/baetyl/utils"
	"os"
	"time"

	"github.com/baetyl/baetyl/logger"
	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
)

func newRecord(data []byte) (*config.Record, error) {
	var r config.Record
	err := json.Unmarshal(data, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

type progress struct {
	event   *EventOTA
	records []*config.Record
}

func newProgress(event *EventOTA) *progress {
	return &progress{
		event:   event,
		records: []*config.Record{},
	}
}

func (p *progress) append(step, msg, err string) {
	p.records = append(p.records, &config.Record{Time: time.Now().UTC().String(), Step: step, Error: err})
}

type operator interface {
	report(...*progress) *config.Inspect
	dying() <-chan struct{}
	clean(string)
}

type otalog struct {
	cfg      config.OTAInfo
	opt      operator
	progress *progress
	rlog     logger.Logger
	log      logger.Logger
}

func newOTALog(cfg config.OTAInfo, opt operator, event *EventOTA, log logger.Logger) *otalog {
	o := &otalog{
		cfg:      cfg,
		opt:      opt,
		progress: newProgress(event),
	}
	if event == nil {
		if !utils.FileExists(cfg.Logger.Path) {
			return nil
		}
		o.rlog = logger.New(cfg.Logger)
		o.log = log
	} else {
		o.rlog = logger.New(cfg.Logger, baetyl.OTAKeyTrace, event.Trace, baetyl.OTAKeyType, event.Type)
		o.log = log.WithField(baetyl.OTAKeyTrace, event.Trace).WithField(baetyl.OTAKeyType, event.Type)
		o.write(baetyl.OTAReceived, "ota event is received", nil)
	}
	return o
}

func (o *otalog) write(step, msg string, err error) {
	if err == nil {
		o.progress.append(step, msg, "")
		o.rlog.WithField(baetyl.OTAKeyStep, step).Infof(msg)
	} else {
		o.progress.append(step, msg, err.Error())
		o.rlog.WithField(baetyl.OTAKeyStep, step).WithError(err).Errorf(msg)
	}
}

func (o *otalog) isFinished() bool {
	l := len(o.progress.records)
	if l == 0 {
		return false
	}
	switch o.progress.records[l-1].Step {
	case baetyl.OTAUpdated, baetyl.OTARolledBack, baetyl.OTAFailure, baetyl.OTATimeout:
		return true
	}
	return false
}

func (o *otalog) isSuccess() bool {
	l := len(o.progress.records)
	if l == 0 {
		return false
	}
	return o.progress.records[l-1].Step == baetyl.OTAUpdated
}

func (o *otalog) load() bool {
	file, err := os.Open(o.cfg.Logger.Path)
	if err != nil {
		o.log.WithError(err).Warnf("failed to open log")
		return false
	}
	defer file.Close()

	records := []*config.Record{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		r, err := newRecord(scanner.Bytes())
		if err != nil {
			o.log.WithError(err).Warnf("failed to parse config.Record")
			return false
		}
		if o.progress.event == nil && r.Trace != "" {
			o.progress.event = &EventOTA{
				Trace: r.Trace,
			}
		}
		records = append(records, r)
	}

	if len(records) == 0 {
		return false
	}

	changed := len(o.progress.records) != len(records)
	o.progress.records = records
	return changed
}

func (o *otalog) wait() {
	o.log.Infof("waiting ota to finish")
	defer o.log.Infof("ota is finished")
	defer o.close()

	ticker := time.NewTicker(time.Second)
	timer := time.NewTimer(o.cfg.Timeout)
	for {
		select {
		case <-ticker.C:
			if o.load() {
				io := o.opt.report(o.progress)
				if o.isFinished() {
					if o.isSuccess() {
						o.opt.clean(io.Software.ConfVersion)
					}
					return
				}
			}
		case <-timer.C:
			o.load()
			if !o.isFinished() {
				o.write(baetyl.OTATimeout, "ota is timed out", nil)
				o.log.Warnf("ota is timed out")
			}
			o.opt.report(o.progress)
			return
		case <-o.opt.dying():
			o.load()
			o.opt.report(o.progress)
			return
		}
	}
}

func (o *otalog) close() {
	if o.isFinished() {
		err := os.RemoveAll(o.cfg.Logger.Path)
		if err != nil {
			o.log.WithError(err).Warnf("failed to remove log")
		}
	}
}
