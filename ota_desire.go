package baetyl

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"

	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/protocol/mqtt"
	"github.com/baetyl/baetyl/utils"

	"github.com/256dpi/gomqtt/packet"
)

type desireContext struct {
	ctx  context.Context
	log  logger.Logger
	rt   *runtime
	disp *mqtt.Dispatcher
}

const (
	desireEventOTA    string = "OTA"
	desireEventUpdate string = "UPDATE"
)

type desireData struct {
	Time    time.Time   `json:"time"`
	Event   string      `json:"event"`
	Payload interface{} `json:"content"`
}

type desireUpdate struct {
	Trace   string       `json:"trace,omitempty"`
	Version string       `json:"version,omitempty"`
	Config  updateVolume `json:"config,omitempty"`
}

func (rt *runtime) runDesire(ctx context.Context) error {
	dc := desireContext{
		ctx: ctx,
		log: rt.log.WithField("ota", "desire"),
		rt:  rt,
	}
	ci := mqtt.ClientInfo{}
	utils.SetDefaults(&ci)
	ci.Subscriptions = []mqtt.TopicInfo{
		mqtt.TopicInfo{
			QOS:   1,
			Topic: fmt.Sprintf("$baidu/iot/edge/%s/core/backward", rt.cfg.Manage.ClientID),
		},
	}
	ci.Address = rt.cfg.Manage.Desire.Address
	ci.Username = rt.cfg.Manage.Desire.Username
	ci.ClientID = rt.cfg.Manage.ClientID
	ci.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{rt.license.cert},
		RootCAs:      rt.license.pool,
	}
	dc.disp = mqtt.NewDispatcher(ci, dc.log)
	err := dc.disp.Start(&dc)
	if err != nil {
		dc.log.Errorf("start mqtt dispatcher fail: %s", err.Error())
		return err
	}
	dc.log.Infoln("begin receiving desire")
	<-ctx.Done()
	dc.disp.Close()
	dc.log.Infoln("receiving desire stopped")
	return ctx.Err()
}

func (dc *desireContext) ProcessPublish(pack *packet.Publish) error {
	dc.log.Infoln("receive desire")
	dt := desireData{}
	err := json.Unmarshal(pack.Message.Payload, &dt)
	if err != nil {
		return fmt.Errorf("invalid desire data: %s", err.Error())
	}
	switch dt.Event {
	case desireEventOTA:
		dt.Payload = &updateData{}
	case desireEventUpdate:
		dt.Payload = &desireUpdate{}
	default:
		return fmt.Errorf("unexpected desire event: %s", dt.Event)
	}
	err = json.Unmarshal(pack.Message.Payload, &dt)
	if err != nil {
		return fmt.Errorf("invalid desire event: %s", err.Error())
	}
	reply := func() {
		if pack.Message.QOS == 1 {
			puback := packet.NewPuback()
			puback.ID = pack.ID
			dc.disp.Send(puback)
		}
	}
	switch payload := dt.Payload.(type) {
	case *updateData:
		err = dc.rt.remoteUpdate(dc.ctx, payload, reply)
	case *desireUpdate:
		err = dc.rt.remoteUpdate(dc.ctx, &updateData{
			Type:    "",
			Trace:   payload.Trace,
			Version: payload.Version,
			Config:  payload.Config,
		}, reply)
	default:
		reply()
	}
	if err != nil {
		dc.log.Errorf("ota desire fail: %s", err.Error())
	}
	return nil
}

func (dc *desireContext) ProcessPuback(*packet.Puback) error {
	return nil
}

func (dc *desireContext) ProcessError(err error) {
	dc.log.Errorf("%s", err.Error())
}
