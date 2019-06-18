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
