// Package dm 设备管理实现
package dm

import (
	"encoding/json"
	"fmt"

	dmctx "github.com/baetyl/baetyl-go/v2/dmcontext"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/mqtt"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	bh "github.com/timshannon/bolthold"
	bolt "go.etcd.io/bbolt"
)

const (
	ReportTopicTemplate          = "thing/%s/%s/property/post"
	DeltaTopicTemplate           = "thing/%s/%s/property/invoke"
	GetTopicTemplate             = "$baetyl/device/%s/get"
	GetResponseTopicTemplate     = "$baetyl/device/%s/getResponse"
	EventTopicTemplate           = "thing/%s/%s/raw/c2d"
	EventReportTopicTemplate     = "thing/%s/%s/event/post"
	PropertyGetTopicTemplate     = "thing/%s/%s/property/get"
	LifecycleReportTopicTemplate = "thing/%s/%s/lifecycle/post"
	DefaultMQTTQOS               = 1
)

func initDeviceTopic(name, version, deviceModel string) dmctx.DeviceInfo {
	return dmctx.DeviceInfo{
		Name:        name,
		Version:     version,
		DeviceModel: deviceModel,
		DeviceTopic: dmctx.DeviceTopic{
			Delta:           mqtt.QOSTopic{QOS: DefaultMQTTQOS, Topic: fmt.Sprintf(DeltaTopicTemplate, deviceModel, name)},
			Report:          mqtt.QOSTopic{QOS: DefaultMQTTQOS, Topic: fmt.Sprintf(ReportTopicTemplate, deviceModel, name)},
			Event:           mqtt.QOSTopic{QOS: DefaultMQTTQOS, Topic: fmt.Sprintf(EventTopicTemplate, deviceModel, name)},
			Get:             mqtt.QOSTopic{QOS: DefaultMQTTQOS, Topic: fmt.Sprintf(GetTopicTemplate, name)},
			GetResponse:     mqtt.QOSTopic{QOS: DefaultMQTTQOS, Topic: fmt.Sprintf(GetResponseTopicTemplate, name)},
			EventReport:     mqtt.QOSTopic{QOS: DefaultMQTTQOS, Topic: fmt.Sprintf(EventReportTopicTemplate, deviceModel, name)},
			PropertyGet:     mqtt.QOSTopic{QOS: DefaultMQTTQOS, Topic: fmt.Sprintf(PropertyGetTopicTemplate, deviceModel, name)},
			LifecycleReport: mqtt.QOSTopic{QOS: DefaultMQTTQOS, Topic: fmt.Sprintf(LifecycleReportTopicTemplate, deviceModel, name)},
		},
	}
}

func (dm *deviceManager) initStore() error {
	return dm.store.Bolt().Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(dm.id)
		return errors.Trace(err)
	})
}

func (dm *deviceManager) getDevice(name string) (d *device, err error) {
	err = dm.store.Bolt().View(func(tx *bolt.Tx) error {
		b := tx.Bucket(dm.id)
		prev := b.Get([]byte(name))
		if len(prev) == 0 {
			return errors.Trace(bh.ErrNotFound)
		}
		d = &device{}
		return errors.Trace(json.Unmarshal(prev, d))
	})
	return
}

func (dm *deviceManager) createDevice(name string, d *device) (err error) {
	return dm.store.Bolt().Batch(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(dm.id)
		if err != nil {
			return errors.Trace(err)
		}

		data := b.Get([]byte(name))
		if len(data) != 0 {
			return errors.Trace(bh.ErrKeyExists)
		}
		data, err = json.Marshal(d)
		if err != nil {
			return errors.Trace(err)
		}

		return errors.Trace(b.Put([]byte(name), data))
	})
}

func (dm *deviceManager) updateDevice(name string, d *device) error {
	return dm.store.Bolt().Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(dm.id)
		if err != nil {
			return errors.Trace(err)
		}
		data, err := json.Marshal(d)
		if err != nil {
			return errors.Trace(err)
		}
		return errors.Trace(b.Put([]byte(name), data))
	})
}

func (dm *deviceManager) delDevice(name string) (err error) {
	return dm.store.Bolt().Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(dm.id)
		return b.Delete([]byte(name))
	})
}

func (dm *deviceManager) deviceDesire(name string, delta v1.Delta) (diff v1.Delta, err error) {
	err = dm.store.Bolt().Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(dm.id)
		prev := bucket.Get([]byte(name))
		if len(prev) == 0 {
			return errors.Trace(bh.ErrNotFound)
		}
		var d device
		err := json.Unmarshal(prev, &d)
		if err != nil {
			return errors.Trace(err)
		}
		if d.Desire == nil {
			d.Desire = v1.Desire(delta)
		} else {
			// nil delta is illegal
			if delta == nil {
				delta = v1.Delta{}
			}
			d.Desire, err = d.Desire.Patch(delta)
			if err != nil {
				return errors.Trace(err)
			}
		}
		curr, err := json.Marshal(d)
		if err != nil {
			return errors.Trace(err)
		}
		if err := bucket.Put([]byte(name), curr); err != nil {
			return errors.Trace(err)
		}
		res, err := d.Desire.Diff(d.Report)
		if err != nil {
			return errors.Trace(err)
		}
		diff = v1.Delta(res)
		return nil
	})
	return
}

func (dm *deviceManager) deviceReport(name string, report v1.Report) (diff v1.Delta, err error) {
	err = dm.store.Bolt().Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(dm.id)
		prev := bucket.Get([]byte(name))
		if len(prev) == 0 {
			return errors.Trace(bh.ErrNotFound)
		}
		var d device
		err := json.Unmarshal(prev, &d)
		if err != nil {
			return errors.Trace(err)
		}
		if d.Report == nil {
			d.Report = report
		} else {
			err = d.Report.Merge(report)
			if err != nil {
				return errors.Trace(err)
			}
		}
		curr, err := json.Marshal(d)
		if err != nil {
			return errors.Trace(err)
		}
		if err := bucket.Put([]byte(name), curr); err != nil {
			return errors.Trace(err)
		}
		res, err := d.Desire.Diff(d.Report)
		if err != nil {
			return errors.Trace(err)
		}
		diff = v1.Delta(res)
		return nil
	})
	return
}

func (dm *deviceManager) filterProperties(d *device, values map[string]interface{}, allows map[string]struct{}) {
	props := map[string]dmctx.DeviceProperty{}
	// TODO check value type
	for _, prop := range d.DeviceModel.Properties {
		props[prop.Name] = prop
	}
	for key := range values {
		_, pExist := props[key]
		_, aExist := allows[key]
		if !pExist && !aExist {
			delete(values, key)
		}
	}
}
