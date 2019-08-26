package main

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewEvent(t *testing.T) {
	e := Event{
		Time: time.Now(),
		Type: Update,
		Content: &UpdateEvent{
			Version: "v2",
		},
	}
	d, err := json.Marshal(e)
	assert.NoError(t, err)
	fmt.Println(string(d))
	got, err := NewEvent(d)
	assert.NoError(t, err)
	u, ok := got.Content.(*UpdateEvent)
	assert.True(t, ok)
	assert.Equal(t, "v2", u.Version)
}

func TestOTAEvent(t *testing.T) {
	raw := `
	{
		"id": 1,
		"event": "OTA",
		"content": {
			"type": "APP",
			"version": "v2",
			"trace": "x-x-x-x",
			"volume": {
				"name": "xxx",
				"path": "sdvfdbv",
				"meta": {
					"url": "cadsvf",
					"md5": "avsdfvgb",
					"version": "v3"
				}
			}
		}
	}`
	event, err := NewEvent([]byte(raw))
	assert.NoError(t, err)
	assert.Equal(t, "OTA", string(event.Type))
	ota, ok := event.Content.(*EventOTA)
	assert.True(t, ok)
	assert.Equal(t, "APP", ota.Type)
	assert.Equal(t, "v2", ota.Version)
	assert.Equal(t, "x-x-x-x", ota.Trace)
	assert.Equal(t, "xxx", ota.Volume.Name)
	assert.Equal(t, "sdvfdbv", ota.Volume.Path)
	assert.Equal(t, "cadsvf", ota.Volume.Meta.URL)
	assert.Equal(t, "avsdfvgb", ota.Volume.Meta.MD5)
	assert.Equal(t, "v3", ota.Volume.Meta.Version)
}
