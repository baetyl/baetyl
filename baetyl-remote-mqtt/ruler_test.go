package main

import (
	"testing"

	"github.com/baetyl/baetyl/protocol/mqtt"
	"github.com/stretchr/testify/assert"
)

func TestDefaults(t *testing.T) {
	rule := new(Rule)
	rule.Hub.Subscriptions = []mqtt.TopicInfo{mqtt.TopicInfo{Topic: "t1"}}
	rule.Remote.Name = "remote"
	rule.Remote.Subscriptions = []mqtt.TopicInfo{mqtt.TopicInfo{Topic: "t2"}}

	hub := new(mqtt.ClientInfo)
	remote := new(mqtt.ClientInfo)
	defaults(rule, hub, remote)
	assert.Equal(t, "remote", hub.ClientID)
	assert.Equal(t, "remote", remote.ClientID)

	hub = new(mqtt.ClientInfo)
	remote = new(mqtt.ClientInfo)
	remote.ClientID = "1"
	defaults(rule, hub, remote)
	assert.Equal(t, "1", hub.ClientID)
	assert.Equal(t, "1", remote.ClientID)

	hub = new(mqtt.ClientInfo)
	hub.ClientID = "2" // ignore
	remote = new(mqtt.ClientInfo)
	remote.ClientID = "1"
	defaults(rule, hub, remote)
	assert.Equal(t, "1", hub.ClientID)
	assert.Equal(t, "1", remote.ClientID)

	hub = new(mqtt.ClientInfo)
	hub.ClientID = "2" // ignore
	remote = new(mqtt.ClientInfo)
	remote.ClientID = "1"
	rule.Hub.ClientID = "3"
	rule.Remote.ClientID = "4"
	defaults(rule, hub, remote)
	assert.Equal(t, "3", hub.ClientID)
	assert.Equal(t, "4", remote.ClientID)
}
