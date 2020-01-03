package main

import (
	"encoding/json"
	"errors"
	"github.com/baetyl/baetyl/baetyl-agent/common"
	"github.com/baetyl/baetyl/baetyl-agent/config"
	"io/ioutil"
	"os"
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
	defer t.Stop()
	err := a.active(nil)
	if err != nil {
		a.ctx.Log().Errorf("active error", err.Error())
	}
	for {
		select {
		case <-t.C:
			err := a.active(nil)
			if err != nil {
				a.ctx.Log().Errorf("active error", err.Error())
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
	a.ctx.Log().Infof("batch uuid = %s", batchUuid)
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
	activation := config.Activation{
		FingerprintValue: fp,
		PenetrateData:    attrs,
	}
	reportInfo := config.ForwardInfo{
		Metadata:   request,
		Activation: activation,
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
	nodeName := res.Metadata["node"].(string)
	ns := res.Metadata["namespace"].(string)
	a.ctx.Log().Debugf("active node name = %s", nodeName)
	a.ctx.Log().Debugf("active namespace = %s", ns)
	if err := os.Setenv(common.NodeName, nodeName); err != nil {
		a.ctx.Log().Errorf("set env node name error", err.Error())
	}
	if err := os.Setenv(common.NodeNamespace, ns); err != nil {
		a.ctx.Log().Errorf("set env node namespace error", err.Error())
	}
	n := &node{
		Name:      nodeName,
		Namespace: ns,
	}
	a.node = n
	a.ctx.Log().Debugf("active set agent node  = %v", *a.node)
	a.ctx.Log().Debugf("active set agent ，point = %p", a)
	a.ctx.Log().Debugf("active set agent = %+v", a)
	a.ctx.Log().Debugf("active delta = %v", res.Delta)
	if len(res.Delta) != 0 {
		le := &EventLink{
			Trace: res.Metadata["trace"].(string),
			Type:  res.Metadata["type"].(string),
		}
		le.Info = res.Delta
		e := &Event{
			Time:    time.Time{},
			Type:    EventType(le.Type),
			Content: le,
		}
		a.ctx.Log().Infof("active event: %+v", *e)
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
