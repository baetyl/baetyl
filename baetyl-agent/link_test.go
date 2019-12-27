package main

import (
	"github.com/goinggo/mapstructure"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetVolumeVersion(t *testing.T) {
	volumePath := "var/db/baetyl/hub-conf/v3"
	version := getVolumeVersion(volumePath)
	assert.Equal(t, version, "v3")
	volumePath = "var/db/baetyl/hub-conf/v3/"
	version = getVolumeVersion(volumePath)
	assert.Equal(t, version, "v3")
}

type TestApp struct {
	AppName    string `yaml:"appName,omitempty"`
	AppVersion string `yaml:"appVersion,omitempty"`
}

func TestTransform(t *testing.T) {
	d := map[string]interface{}{
		"appName":    "hub",
		"appVersion": "32423",
	}
	var app TestApp
	err := mapstructure.Decode(d, &app)
	assert.NoError(t, err)
	expected := TestApp{
		AppName:    "hub",
		AppVersion: "32423",
	}
	assert.Equal(t, app, expected)
}
