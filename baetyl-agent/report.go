package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/baetyl/baetyl/baetyl-agent/common"
	"github.com/baetyl/baetyl/baetyl-agent/config"
	"github.com/baetyl/baetyl/logger"
	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
	"github.com/baetyl/baetyl/utils"
)

type EventLink struct {
	Info  map[string]interface{}
	Trace string
	Type  string
}

func (a *agent) reporting() error {
	t := time.NewTicker(a.cfg.Remote.Report.Interval)
	for {
		select {
		case <-t.C:
			a.report()
		case <-a.tomb.Dying():
			return nil
		}
	}
}

// Report reports info
func (a *agent) report(pgs ...*progress) *config.Inspect {
	defer utils.Trace("report", logger.Debugf)()

	i, err := a.ctx.InspectSystem()
	if err != nil {
		a.ctx.Log().WithError(err).Warnf("failed to config.Inspect stats")
		i = baetyl.NewInspect()
		i.Error = err.Error()
	}
	if utils.FileExists(a.cfg.OTA.Logger.Path) && len(pgs) == 0 {
		return nil
	}

	io := &config.Inspect{Inspect: i}
	for _, pg := range pgs {
		if io.OTA == nil {
			io.OTA = map[string][]*config.Record{}
		}
		if pg.event != nil && pg.event.Trace != "" {
			io.OTA[pg.event.Trace] = pg.records
		}
	}
	if a.mqtt == nil {
		a.ctx.Log().Debugf("report set agent ，point = %p", a)
		a.ctx.Log().Debugf("report set agent = %+v", a)
		currentApp, err := a.getCurrentApp(io.Inspect)
		if err != nil {
			a.ctx.Log().WithError(err).Warnf("failed to get current app info")
			return nil
		}
		info := config.ForwardInfo{
			Status: io,
			Apps:   currentApp,
		}
		if a.node == nil {
			a.ctx.Log().WithError(err).Warnf("node nil , to active")
			actInfo, err := a.collectActiveInfo(i)
			if err != nil {
				a.ctx.Log().WithError(err).Warnf("collect active info error")
				return nil
			}
			info.Activation = *actInfo
		}
		req, err := json.Marshal(info)
		if err != nil {
			a.ctx.Log().WithError(err).Warnf("failed to marshal report info")
			return nil
		}
		var res config.BackwardInfo
		resData, err := a.sendRequest("POST", a.cfg.Remote.Report.URL, req)
		if err != nil {
			a.ctx.Log().WithError(err).Warnf("failed to send report data")
			return nil
		}
		err = json.Unmarshal(resData, &res)
		if err != nil {
			a.ctx.Log().WithError(err).Warnf("error to unmarshal response data returned")
			return nil
		}
		if a.node == nil {
			a.node = &node{
				Name:      res.Metadata[string(common.Node)].(string),
				Namespace: res.Metadata[common.KeyContextNamespace].(string),
			}
		}
		if res.Delta != nil {
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
			select {
			case oe := <-a.events:
				a.ctx.Log().Warnf("discard old event: %+v", *oe)
				a.events <- e
			case a.events <- e:
			}
		}
	} else {
		payload, err := json.Marshal(io)
		if err != nil {
			a.ctx.Log().WithError(err).Warnf("failed to marshal stats")
			return nil
		}
		a.ctx.Log().Debugln("stats", string(payload))
		// TODO: connect with device management on cloud
		// p := packet.NewPublish()
		// p.Message.Topic = a.cfg.Remote.Report.Topic
		// p.Message.Payload = payload
		// err = a.mqtt.Send(p)
		// if err != nil {
		// 	a.ctx.Log().WithError(err).Warnf("failed to report stats by mqtt")
		// }
		err = a.send(payload)
		if err != nil {
			a.ctx.Log().WithError(err).Warnf("failed to report stats by https")
			return nil
		}
	}
	return io
}

func (a *agent) collectActiveInfo(inspect *baetyl.Inspect) (*config.Activation, error) {
	attrs := a.attrs
	if attrs == nil {
		attrs = map[string]string{}
		for _, item := range a.cfg.Attributes {
			attrs[item.Name] = item.Value
		}
	}
	a.ctx.Log().Debugln("active attributes : ", attrs)
	fp := attrs[common.KeyFingerprintValue]
	if a.cfg.Fingerprints != nil && len(a.cfg.Fingerprints) > 0 {
		for _, instance := range a.cfg.Fingerprints {
			if instance.Proof == common.Input {
				fp = attrs[os.Getenv(common.BatchInputField)]
				break
			}
			proof, err := collectFP(instance.Proof, instance.Value, inspect)
			if err != nil {
				return nil, err
			}
			if proof != "" {
				fp = proof
				break
			}
		}
	}
	if fp == "" {
		return nil, errors.New("cannot get fingerprint")
	}
	return &config.Activation{
		FingerprintValue: fp,
		PenetrateData:    attrs,
	}, nil
}

func (a *agent) send(data []byte) error {
	body, key, err := a.encryptData(data)
	if err != nil {
		return err
	}
	header := map[string]string{
		"x-iot-edge-sn":       a.certSN,
		"x-iot-edge-key":      key,
		"x-iot-edge-clientid": a.cfg.Remote.MQTT.ClientID,
		"Content-Type":        "application/x-www-form-urlencoded",
	}
	_, err = a.http.SendPath("POST", a.cfg.Remote.Report.URL, body, header)
	return err
}

func (a *agent) encryptData(data []byte) ([]byte, string, error) {
	aesKey := utils.NewAesKey()
	// encrypt data using AES
	body, err := utils.AesEncrypt(data, aesKey)
	if err != nil {
		return nil, "", err
	}
	// encrypt AES key using RSA
	k, err := utils.RsaPrivateEncrypt(aesKey, a.certKey)
	if err != nil {
		return nil, "", err
	}
	// encode key using BASE64
	key := base64.StdEncoding.EncodeToString(k)
	// encode body using BASE64
	body = []byte(base64.StdEncoding.EncodeToString(body))
	return body, key, nil
}

func collectFP(proof common.Proof, value string, inspect *baetyl.Inspect) (string, error) {
	switch proof {
	case common.HostID:
		return inspect.Hardware.HostInfo.HostID, nil
	case common.CPU:
		return collectCPU(inspect)
	case common.MAC:
		return collectMAC(value, inspect)
	case common.SN:
		return collectSN(value)
	default:
		return "", errors.New("proof invalid")
	}
}

// collectCPU get cpu info
func collectCPU(inspect *baetyl.Inspect) (string, error) {
	// todo collect CPU info
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
	snByte, err := ioutil.ReadFile(path.Join(common.SNPath, value))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(snByte)), nil
}
