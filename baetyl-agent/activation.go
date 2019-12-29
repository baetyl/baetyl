package main

import (
	"encoding/json"
	"errors"
	"github.com/baetyl/baetyl/baetyl-agent/common"
	"github.com/baetyl/baetyl/baetyl-agent/config"
	"io/ioutil"
	"path"
	"strings"
	"time"

	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
)

const (
	batchUuid = "batchuuid"

	proofDiskID = "diskID"
	proofHostID = "hostID"
	proofCPU    = "cpu"
	proofMAC    = "mac"
	proofSN     = "sn"

	snPath = "var/lib/baetyl/sn/"
)

func (a *agent) autoActive() error {
	t := time.NewTicker(a.cfg.Interval)
	err := a.active(nil)
	if err != nil {
		a.ctx.Log().WithError(err)
	} else {
		t.Stop()
	}
	for {
		select {
		case <-t.C:
			err := a.active(nil)
			if err != nil {
				a.ctx.Log().WithError(err)
			} else {
				t.Stop()
			}
		case <-a.tomb.Dying():
			return nil
		}
	}
}

func (a *agent) active(attrs map[string]string) (err error) {
	if attrs == nil {
		attrs = map[string]string{}
		for _, item := range a.cfg.Attributes {
			attrs[item.Name] = item.Value
		}
	}
	a.ctx.Log().Debugln("info c attributes : ", attrs)
	batchUuid, ok := attrs[batchUuid]
	if !ok {
		return errors.New("batch uuid can't be null")
	}
	a.ctx.Log().Infof("batch uuid", batchUuid)
	inspect, err := a.ctx.InspectSystem()
	if err != nil {
		return err
	}
	var fp string
	for _, instance := range a.cfg.Fingerprints {
		proof, err := collectFP(instance.Proof, instance.Value, inspect)
		if err != nil {
			return err
		}
		fp += proof + ","
	}
	if fp == "" {
		return errors.New("cannot get fingerprint")
	}
	request := map[string]string{
		common.ResourceType: string(common.Batch),
		common.ResourceName: batchUuid,
	}
	activeData := config.ActiveData{
		FingerprintValue: fp,
		PenetrateData:    attrs,
	}
	reportInfo := config.ForwardInfo{
		Request:    request,
		ActiveData: activeData,
	}

	var res config.BackwardInfo
	resData, err := a.sendData(reportInfo)
	if err != nil {
		a.ctx.Log().WithError(err).Warnf("failed to send report data by link")
		return nil
	}
	err = json.Unmarshal(resData, &res)
	if err != nil {
		a.ctx.Log().WithError(err).Warnf("error to unmarshal response data returned by link")
		return nil
	}
	if len(res.Delta) != 0 {
		le := &EventLink{
			Trace: res.Response["trace"].(string),
			Type:  res.Response["type"].(string),
		}
		le.Info = res.Delta
		e := &Event{
			Time:    time.Time{},
			Type:    EventType(le.Type),
			Content: le,
		}
		select {
		case a.events <- e:
		default:
			a.ctx.Log().Warnf("discard event: %+v", *e)
		}
	}
	return nil
}

func collectFP(proof, value string, inspect *baetyl.Inspect) (string, error) {
	switch proof {
	case proofHostID:
		return inspect.Hardware.HostInfo.HostID, nil
	case proofCPU:
		return collectCPU(inspect)
	case proofMAC:
		return collectMAC(value, inspect)
	case proofSN:
		return collectSN(value)
	default:
		return "", errors.New("proof invalid")
	}
}

// collectCPU get cpu info
func collectCPU(inspect *baetyl.Inspect) (string, error) {
	// todo
	return "", nil
}

// collectMAC get mac address
func collectMAC(value string, inspect *baetyl.Inspect) (string, error) {
	interfStat := inspect.Hardware.NetInfo.Interfaces
	var mac string
	for _, interf := range interfStat {
		if interf.Name == value {
			mac = interf.MAC
			break
		}
	}
	return mac, nil
}

// collectSN get sn from var/db/baetyl/data/
func collectSN(value string) (string, error) {
	snByte, err := ioutil.ReadFile(path.Join(snPath, value))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(snByte)), nil
}
