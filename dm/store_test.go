// Package dm 设备管理实现
package dm

import (
	"fmt"
	"io/ioutil"
	"testing"

	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl/v2/store"
	"github.com/stretchr/testify/assert"
)

func genDeviceManager(t *testing.T) *deviceManager {
	f, err := ioutil.TempFile("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())

	s, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, s)
	dm := &deviceManager{
		id:    []byte("baetyl-edge-device"),
		store: s,
	}
	return dm
}

func TestDevice(t *testing.T) {
	name := "device-1"
	dm := genDeviceManager(t)
	err := dm.initStore()
	assert.NoError(t, err)
	err = dm.createDevice(name, &device{})
	assert.NoError(t, err)

	// each test case depend on the previous device states
	tests := []struct {
		report v1.Report
		rDiff  v1.Delta
		rErr   error
		delta  v1.Delta
		dDiff  v1.Delta
		dErr   error
		d      *device
		err    error
	}{
		{
			report: nil,
			rDiff:  v1.Delta{},
			rErr:   nil,
			delta:  nil,
			dDiff:  v1.Delta{},
			dErr:   nil,
			d:      &device{},
			err:    nil,
		},
		{
			report: v1.Report{},
			rDiff:  v1.Delta{},
			rErr:   nil,
			delta:  v1.Delta{},
			dDiff:  v1.Delta{},
			dErr:   nil,
			d:      &device{Report: v1.Report{}, Desire: v1.Desire{}},
			err:    nil,
		},
		{
			report: v1.Report{"a": "1"},
			rDiff:  v1.Delta{},
			rErr:   nil,
			delta:  nil,
			dDiff:  v1.Delta{},
			dErr:   nil,
			d:      &device{Report: v1.Report{"a": "1"}, Desire: v1.Desire{}},
			err:    nil,
		},
		{
			report: nil,
			rDiff:  v1.Delta{},
			rErr:   nil,
			delta:  v1.Delta{"a": "1"},
			dDiff:  v1.Delta{},
			dErr:   nil,
			d:      &device{Report: v1.Report{"a": "1"}, Desire: v1.Desire{"a": "1"}},
			err:    nil,
		},
		{
			report: nil,
			rDiff:  v1.Delta{},
			rErr:   nil,
			delta:  v1.Delta{"a": "2"},
			dDiff:  v1.Delta{"a": "2"},
			dErr:   nil,
			d:      &device{Report: v1.Report{"a": "1"}, Desire: v1.Desire{"a": "2"}},
			err:    nil,
		},
		{
			report: v1.Report{"a": "2", "b": "2"},
			rDiff:  v1.Delta{},
			rErr:   nil,
			delta:  v1.Delta{"b": "2"},
			dDiff:  v1.Delta{},
			dErr:   nil,
			d:      &device{Report: v1.Report{"a": "2", "b": "2"}, Desire: v1.Desire{"a": "2", "b": "2"}},
			err:    nil,
		},
	}
	var res v1.Delta
	var dev *device
	for _, tt := range tests {
		res, err = dm.deviceReport(name, tt.report)
		assert.Equal(t, tt.rErr, err)
		assert.Equal(t, tt.rDiff, res)
		res, err = dm.deviceDesire(name, tt.delta)
		assert.Equal(t, tt.dErr, err)
		assert.Equal(t, tt.dDiff, res)
		dev, err = dm.getDevice(name)
		assert.Equal(t, tt.err, err)
		assertEqualDevice(t, tt.d, dev)
	}
	err = dm.delDevice(name)
	assert.NoError(t, err)
}

func assertEqualDevice(t *testing.T, d1 *device, d2 *device) {
	assert.Equal(t, d1.Report, d2.Report)
	assert.Equal(t, d1.Desire, d2.Desire)
}
