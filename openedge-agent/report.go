package main

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/baidu/openedge/logger"
	openedge "github.com/baidu/openedge/sdk/openedge-go"
	"github.com/baidu/openedge/utils"
)

type inspect struct {
	*openedge.Inspect `json:",inline"`
	OTA               map[string][]*record `json:"ota,omitempty"`
}

func (a *agent) reporting() error {
	t := time.NewTicker(a.cfg.Remote.Report.Interval)
	a.report()
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
func (a *agent) report(pgs ...*progress) *inspect {
	defer utils.Trace("report", logger.Debugf)()

	i, err := a.ctx.InspectSystem()
	if err != nil {
		a.ctx.Log().WithError(err).Warnf("failed to inspect stats")
		i = openedge.NewInspect()
		i.Error = err.Error()
	}

	io := &inspect{Inspect: i}
	for _, pg := range pgs {
		if io.OTA == nil {
			io.OTA = map[string][]*record{}
		}
		if pg.event != nil && pg.event.Trace != "" {
			io.OTA[pg.event.Trace] = pg.records
		}
	}

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
	return io
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
