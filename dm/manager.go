// Package dm 设备管理实现
package dm

import (
	"encoding/json"
	"io"
	gosync "sync"
	"time"

	"github.com/baetyl/baetyl-go/v2/context"
	dmctx "github.com/baetyl/baetyl-go/v2/dmcontext"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/mqtt"
	mqtt2 "github.com/baetyl/baetyl-go/v2/mqtt"
	v2plugin "github.com/baetyl/baetyl-go/v2/plugin"
	"github.com/baetyl/baetyl-go/v2/pubsub"
	"github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
	bh "github.com/timshannon/bolthold"

	"github.com/baetyl/baetyl/v2/config"
	"github.com/baetyl/baetyl/v2/node"
	"github.com/baetyl/baetyl/v2/plugin"
	"github.com/baetyl/baetyl/v2/sync"
)

const (
	KeyName                      = "name"
	KeyVersion                   = "version"
	KeyStatus                    = "status"
	DefaultDeviceManagerClientID = "baetyl-dm-client"
	DefaultDeviceStoreID         = "baetyl-edge-device"
)

var (
	ErrInvalidPubsubPlugin = errors.New("invalid pubsub plugin")
	ErrMqttClientNotExist  = errors.New("mqtt client not exist")
	ErrDeviceNotExist      = errors.New("device not exist")
	ErrInvalidReport       = errors.New("invalid report")
	ErrMissingField        = errors.New("missing field")
	ErrDeviceModelNotExist = errors.New("device model not exist")
	SystemAllowProperties  = map[string]struct{}{KeyStatus: {}}
)

type DeviceManager interface {
	Start() error
	io.Closer
}

type deviceManager struct {
	id        []byte
	ctx       context.Context
	pb        plugin.Pubsub
	syn       sync.Sync
	processor pubsub.Processor
	store     *bh.Store
	node      node.Node
	mqtt      *mqtt.Client
	lock      gosync.Mutex
	observer  mqtt.Observer
	log       *log.Logger
	tomb      utils.Tomb
	upMsgCh   chan *v1.Message
	downMsgCh chan *v1.Message
	msg       dmctx.Msg
	devices   map[string]dmctx.DeviceInfo
}

type device struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	DeviceModel sync.DeviceModel       `json:"deviceModel"`
	Report      v1.Report              `json:"report"`
	Desire      v1.Desire              `json:"desire"`
	MetaData    map[string]interface{} `json:"metaData"`
	CreateTime  time.Time              `json:"createTime"`
}

func NewDeviceManager(ctx context.Context, store *bh.Store, node node.Node, syn sync.Sync, cfg config.Config) (DeviceManager, error) {
	lg := log.With(log.Any("core", "device manager"))
	pl, err := v2plugin.GetPlugin(cfg.Plugin.Pubsub)
	if err != nil {
		return nil, errors.Trace(err)
	}
	pb, ok := pl.(pubsub.Pubsub)
	if !ok {
		return nil, errors.Trace(ErrInvalidPubsubPlugin)
	}
	dm := &deviceManager{
		ctx: ctx,
		id:  []byte(DefaultDeviceStoreID),
		pb:  pl.(plugin.Pubsub),
		// TODO init by config
		upMsgCh:   make(chan *v1.Message, 1024),
		downMsgCh: make(chan *v1.Message, 1024),
		msg:       dmctx.InitMsg(dmctx.Blink),
		log:       lg,
		store:     store,
		node:      node,
		syn:       syn,
		devices:   make(map[string]dmctx.DeviceInfo),
	}
	if err = dm.initStore(); err != nil {
		return nil, err
	}
	nod, err := node.Get()
	if err != nil {
		return nil, err
	}
	infos, _ := nod.Report[v1.KeyDevices].([]interface{})
	if len(infos) != 0 {
		var comparisonDevArray []string
		comparisonDevices := map[string]string{}
		for _, info := range infos {
			dev, ok := info.(map[string]interface{})
			if ok {
				name, nameExist := dev[KeyName]
				ver, verExist := dev[KeyVersion]
				if verExist && nameExist {
					name, nameOk := name.(string)
					ver, verOk := ver.(string)
					if nameOk && verOk {
						comparisonDevices[name] = ver
						comparisonDevArray = append(comparisonDevArray, name)
					} else {
						dm.log.Error("convert parameters to string error")
					}
				} else {
					dm.log.Error("name or version not exist in map")
				}
			} else {
				dm.log.Error("convert parameters to map error")
			}
		}
		deviceModelMap, err := dm.getDeviceModels(comparisonDevArray)
		if err != nil {
			return nil, errors.Trace(err)
		}
		for devName, model := range deviceModelMap {
			if ver, ok := comparisonDevices[devName]; ok {
				devInfo := initDeviceTopic(devName, ver, model.Name)
				dm.devices[devName] = devInfo
			}
		}
		dm.log.Debug("all devices", log.Any("device", dm.devices))
	}
	var subs []mqtt2.QOSTopic
	for _, dev := range dm.devices {
		subs = append(subs, dev.DeviceTopic.Get, dev.DeviceTopic.LifecycleReport)
	}
	dm.observer = newObserver(dm.downMsgCh, dm.log)
	mqttConfig, err := ctx.NewSystemBrokerClientConfig()
	if err != nil {
		return nil, err
	}
	mqttConfig.ClientID = DefaultDeviceManagerClientID
	mqttConfig.Subscriptions = subs
	dm.mqtt, err = ctx.NewBrokerClient(mqttConfig)
	if err != nil {
		return nil, errors.Trace(err)
	}
	handle := newHandler(dm.upMsgCh, dm.log)
	ch, err := pb.Subscribe(sync.TopicDM)
	if err != nil {
		return nil, errors.Trace(err)
	}
	dm.processor = pubsub.NewProcessor(ch, 0, handle)
	return dm, nil
}

func (dm *deviceManager) Start() error {
	dm.log.Debug("starting device manager")
	if err := dm.mqtt.Start(dm.observer); err != nil {
		return errors.Trace(err)
	}
	dm.processor.Start()
	return dm.tomb.Go(dm.processingUpMsg, dm.processingDownMsg)
}

func (dm *deviceManager) processingUpMsg() error {
	for {
		select {
		case <-dm.tomb.Dying():
			return nil
		case msg := <-dm.upMsgCh:
			bytes, err := json.Marshal(msg)
			if err != nil {
				dm.log.Error("failed to marshal msg")
			}
			msgStr := string(bytes)
			dm.log.Debug("receive upside message", log.Any("message", msgStr))
			switch msg.Kind {
			case v1.MessageDevices:
				err = dm.processUpDeviceMsg(msg)
				if err != nil {
					dm.log.Error("failed to process device message", log.Any("message", msgStr), log.Error(err))
				}
			case v1.MessageDeviceDelta:
				err = dm.processUpDeviceDeltaMsg(msg)
				if err != nil {
					dm.log.Error("failed to process device property message", log.Any("message", msgStr), log.Error(err))
				}
			case v1.MessageDeviceEvent:
				err = dm.processUpDeviceEventMsg(msg)
				if err != nil {
					dm.log.Error("failed to process device event message", log.Any("message", msgStr), log.Error(err))
				}
			case v1.MessageDevicePropertyGet:
				err = dm.processUpDevicePropertyGetMsg(msg)
				if err != nil {
					dm.log.Error("failed to process device latest property message", log.Any("message", msgStr), log.Error(err))
				}
			default:
				dm.log.Error("message kind unsupported", log.Any("message", msgStr))
			}
			dm.log.Debug("process upside message successfully", log.Any("message", msgStr))
		}
	}
}

func (dm *deviceManager) processingDownMsg() error {
	for {
		select {
		case <-dm.tomb.Dying():
			return nil
		case msg := <-dm.downMsgCh:
			bytes, err := json.Marshal(msg)
			if err != nil {
				dm.log.Error("failed to marshal msg")
			}
			msgStr := string(bytes)
			dm.log.Debug("receive downside message", log.Any("message", msgStr))
			switch msg.Kind {
			case v1.MessageDeviceDesire:
				err = dm.processDownGetMsg(msg)
				if err != nil {
					dm.log.Error("failed to process device property message", log.Any("message", msgStr), log.Error(err))
				}
			case v1.MessageDeviceReport:
				err = dm.processDownReportMsg(msg)
				if err != nil {
					dm.log.Error("failed to process device report message", log.Any("message", msgStr), log.Error(err))
				}
			case v1.MessageDeviceLifecycleReport:
				err = dm.processDownLifecycleReportMsg(msg)
				if err != nil {
					dm.log.Error("failed to process device lifecycle report message", log.Any("message", msgStr), log.Error(err))
				}
			default:
				dm.log.Error("message kind unsupported", log.Any("message", msgStr))
			}
			dm.log.Debug("process downside message successfully", log.Any("message", msgStr))
		}
	}
}

func (dm *deviceManager) refreshDevice(name string) error {
	d, err := dm.getDevice(name)
	if err != nil {
		return err
	}
	dmMap, err := dm.syn.SyncDeviceModels(name)
	if err != nil {
		return err
	}
	// TODO batch update device
	if devModel, ok := dmMap[name]; ok {
		if devModel.Version == d.DeviceModel.Version {
			return nil
		}
	}
	d.DeviceModel = dmMap[name]
	d.Report = v1.Report{}
	d.Desire = v1.Desire{}
	return dm.updateDevice(name, d)
}

func (dm *deviceManager) addDevice(name string, deviceModel sync.DeviceModel) error {
	d := &device{
		DeviceModel: deviceModel,
		Report:      v1.Report{},
		Desire:      v1.Desire{},
		CreateTime:  time.Now().UTC(),
	}
	if err := dm.createDevice(name, d); err != nil && errors.Cause(err) != bh.ErrKeyExists {
		return err
	}
	return nil
}

// TODO compatible with old cloud computing
func (dm *deviceManager) processUpDeviceMsg(msg *v1.Message) error {
	dm.log.Debug("receive upside message", log.Any("message", msg))
	infos, _ := msg.Content.Value.([]interface{})
	newDevices := map[string]string{}
	comparisonDevices := map[string]string{}
	var comparisonDevArray []string
	var changed bool
	for _, info := range infos {
		dev, ok := info.(map[string]interface{})
		if ok {
			name, nameExist := dev[KeyName]
			ver, verExist := dev[KeyVersion]
			if nameExist && verExist {
				name, nameOk := name.(string)
				ver, verOk := ver.(string)
				if nameOk && verOk {
					newDevices[name] = ver
					if oldDev, ok := dm.devices[name]; !ok {
						comparisonDevices[name] = ver
						comparisonDevArray = append(comparisonDevArray, name)
						changed = true
					} else if oldDev.Version != ver {
						// update device
						oldDev.Version = ver
						if err := dm.refreshDevice(name); err != nil {
							return err
						}
						dm.devices[name] = oldDev
					}
				} else {
					dm.log.Error("convert parameters to string error")
				}
			} else {
				dm.log.Error("name or version not exist in map")
			}
		} else {
			dm.log.Error("convert parameters to map error")
		}
	}
	if len(comparisonDevArray) != 0 {
		dm.log.Debug("start obtaining models for all sub devices")
		deviceModelMap, err := dm.getDeviceModels(comparisonDevArray)
		if err != nil {
			return errors.Trace(err)
		}
		dm.log.Debug("obtaining all sub device models")
		for devName, model := range deviceModelMap {
			if ver, ok := comparisonDevices[devName]; ok {
				devInfo := initDeviceTopic(devName, ver, model.Name)
				dm.devices[devName] = devInfo
				if err := dm.addDevice(devName, model); err != nil {
					return err
				}
			}
		}
		comparisonDevArray = comparisonDevArray[:0]
	}

	// delete device
	if len(newDevices) < len(dm.devices) {
		changed = true
		for name := range dm.devices {
			if _, ok := newDevices[name]; !ok {
				delete(dm.devices, name)
				if err := dm.delDevice(name); err != nil {
					return err
				}
			}
		}
	}
	if changed {
		// TODO avoid useless resubscribe
		dm.log.Info("refresh all sub device topics")
		if err := dm.reSubscribe(); err != nil {
			dm.log.Error("failed to resubscribe", log.Error(err))
		}
	}
	if _, err := dm.node.Report(v1.Report{v1.KeyDevices: infos}, false); err != nil {
		return err
	}
	return nil
}

// processUpDeviceDeltaMsg will not be called.
func (dm *deviceManager) processUpDeviceDeltaMsg(msg *v1.Message) error {
	deviceName := msg.Metadata[dmctx.KeyDevice]
	if info, ok := dm.devices[deviceName]; ok {
		var delta v1.Delta
		if err := msg.Content.Unmarshal(&delta); err != nil {
			return errors.Trace(err)
		}
		if delta == nil {
			return nil
		}
		device, err := dm.getDevice(msg.Metadata[dmctx.KeyDevice])
		if err != nil {
			return errors.Trace(err)
		}
		dm.filterProperties(device, delta, SystemAllowProperties)
		diff, err := dm.deviceDesire(deviceName, delta)
		if err != nil {
			if errors.Cause(err) == bh.ErrKeyExists {
				dm.log.Warn("device not exist and msg will be discard", log.Any("device", device))
				return nil
			}
			return errors.Trace(err)
		}
		if len(diff) == 0 {
			return nil
		}

		msg.Content = dm.msg.GenDeltaBlinkData(delta)
		pld, err := json.Marshal(msg)
		if err != nil {
			return errors.Trace(err)
		}
		dm.lock.Lock()
		if err = dm.mqtt.Publish(mqtt2.QOS(info.DeviceTopic.Delta.QOS),
			info.DeviceTopic.Delta.Topic, pld, 0, false, false); err != nil {
			return errors.Trace(err)
		}
		dm.lock.Unlock()
	}
	return nil
}

func (dm *deviceManager) processUpDeviceEventMsg(msg *v1.Message) error {
	device := msg.Metadata[dmctx.KeyDevice]
	if info, ok := dm.devices[device]; ok {
		pld, err := json.Marshal(msg)
		if err != nil {
			return errors.Trace(err)
		}
		dm.lock.Lock()
		if err := dm.mqtt.Publish(mqtt2.QOS(info.DeviceTopic.Event.QOS),
			info.DeviceTopic.Event.Topic, pld, 0, false, false); err != nil {
			return errors.Trace(err)
		}
		dm.lock.Unlock()
	} else {
		return ErrDeviceNotExist
	}
	return nil
}

func (dm *deviceManager) processUpDevicePropertyGetMsg(msg *v1.Message) error {
	device := msg.Metadata[dmctx.KeyDevice]
	if info, ok := dm.devices[device]; ok {
		var propertyGet dmctx.PropertyGet
		if err := msg.Content.Unmarshal(&propertyGet); err != nil {
			return errors.Trace(err)
		}
		metadata, err := dm.genMetadata(info)
		if err != nil {
			return errors.Trace(err)
		}
		msg.Metadata = metadata
		msg.Content = dm.msg.GenPropertyGetBlinkData(propertyGet.Properties)

		pld, err := json.Marshal(msg)
		if err != nil {
			return errors.Trace(err)
		}
		if err := dm.mqtt.Publish(mqtt2.QOS(info.DeviceTopic.PropertyGet.QOS),
			info.DeviceTopic.PropertyGet.Topic, pld, 0, false, false); err != nil {
			return errors.Trace(err)
		}
	} else {
		return ErrDeviceNotExist
	}
	return nil
}

func (dm *deviceManager) processDownGetMsg(msg *v1.Message) error {
	if device, ok := dm.devices[msg.Metadata[dmctx.KeyDevice]]; ok {
		dev, err := dm.getDevice(device.Name)
		if err != nil {
			return errors.Trace(err)
		}
		deviceShadow := dmctx.DeviceShadow{Name: device.Name, Report: dev.Report, Desire: dev.Desire}
		msg.Kind = v1.MessageResponse
		msg.Content = v1.LazyValue{Value: deviceShadow}
		pld, err := json.Marshal(msg)
		if err != nil {
			return errors.Trace(err)
		}
		dm.lock.Lock()
		if err = dm.mqtt.Publish(mqtt.QOS(device.DeviceTopic.GetResponse.QOS),
			device.DeviceTopic.GetResponse.Topic, pld, 0, false, false); err != nil {
			return errors.Trace(err)
		}
		dm.lock.Unlock()
	} else {
		return ErrDeviceNotExist
	}
	return nil
}

func (dm *deviceManager) processDownReportMsg(msg *v1.Message) error {
	var blinkContent dmctx.ContentBlink
	if err := msg.Content.Unmarshal(&blinkContent); err != nil {
		return errors.Trace(err)
	}
	r, ok := blinkContent.Blink.Properties.(map[string]interface{})
	if !ok {
		return errors.Trace(ErrInvalidReport)
	}
	dev, err := dm.getDevice(msg.Metadata[dmctx.KeyDevice])
	if err != nil {
		return errors.Trace(err)
	}
	dm.filterProperties(dev, r, SystemAllowProperties)
	if len(r) == 0 {
		return nil
	}
	if _, err = dm.deviceReport(msg.Metadata[dmctx.KeyDevice], r); err != nil {
		if errors.Cause(err) == bh.ErrKeyExists {
			dm.log.Warn("device not exist and msg will be discard", log.Any("device", msg.Metadata[dmctx.KeyDevice]))
			return nil
		}
		return errors.Trace(err)
	}
	dev, err = dm.getDevice(msg.Metadata[dmctx.KeyDevice])
	if err != nil {
		return errors.Trace(err)
	}
	msg.Content = v1.LazyValue{Value: dev.Report}
	// it's report to cloud periodic currently,
	// might be better to get delta and decide whether to report in future
	// but it depends on cloud pushing delta to edge
	if err = dm.pb.Publish(sync.TopicUpside, msg); err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (dm *deviceManager) processDownLifecycleReportMsg(msg *v1.Message) error {
	var blinkContent dmctx.ContentBlink
	if err := msg.Content.Unmarshal(&blinkContent); err != nil {
		return errors.Trace(err)
	}
	online, ok := blinkContent.Blink.Params[dmctx.KeyOnlineState]
	if !ok {
		return errors.Trace(ErrMissingField)
	}
	msg.Content = v1.LazyValue{Value: v1.Report{dmctx.KeyOnlineState: online}}
	if err := dm.pb.Publish(sync.TopicUpside, msg); err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (dm *deviceManager) reSubscribe() error {
	dm.lock.Lock()
	defer dm.lock.Unlock()
	if dm.mqtt != nil {
		if err := dm.mqtt.Close(); err != nil {
			return errors.Trace(err)
		}
	} else {
		return errors.Trace(ErrMqttClientNotExist)
	}
	var subs []mqtt2.QOSTopic
	for _, dev := range dm.devices {
		subs = append(subs, dev.DeviceTopic.Get, dev.DeviceTopic.LifecycleReport)
	}
	mqttConfig, err := dm.ctx.NewSystemBrokerClientConfig()
	if err != nil {
		return err
	}
	mqttConfig.ClientID = DefaultDeviceManagerClientID
	mqttConfig.Subscriptions = subs
	dm.mqtt, err = dm.ctx.NewBrokerClient(mqttConfig)
	if err != nil {
		return errors.Trace(err)
	}
	if err := dm.mqtt.Start(dm.observer); err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (dm *deviceManager) Close() error {
	dm.tomb.Kill(nil)
	dm.tomb.Wait()
	dm.processor.Close()
	if dm.mqtt != nil {
		dm.mqtt.Close()
	}
	if dm.store != nil {
		dm.store.Close()
	}
	return nil
}

func (dm *deviceManager) getDeviceModels(names []string) (map[string]sync.DeviceModel, error) {
	deviceModels, err := dm.syn.SyncDeviceModels(names...)
	if err != nil {
		return nil, errors.Trace(err)
	}
	if len(deviceModels) == 0 {
		return nil, ErrDeviceModelNotExist
	}
	return deviceModels, nil
}

func (dm *deviceManager) genMetadata(info dmctx.DeviceInfo) (map[string]string, error) {
	return map[string]string{
		dmctx.KeyDevice:        info.Name,
		dmctx.KeyDeviceProduct: info.DeviceModel,
		dmctx.KeyNode:          dm.ctx.NodeName(),
		dmctx.KeyNodeProduct:   dmctx.NodeProduct,
	}, nil
}
