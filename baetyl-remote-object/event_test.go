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
		Type: Upload,
		Content: &UploadEvent{
			RemotePath: "test",
			LocalPath:  "aaa",
			Zip:        true,
			Meta:       map[string]string{"x": "x1"},
		},
	}
	d, err := json.Marshal(e)
	assert.NoError(t, err)
	fmt.Println(string(d))
	got, err := NewEvent(d)
	assert.NoError(t, err)
	u, ok := got.Content.(*UploadEvent)
	assert.True(t, ok)
	assert.Equal(t, "test", u.RemotePath)
	assert.Equal(t, "aaa", u.LocalPath)
	assert.True(t, u.Zip)
	assert.Equal(t, "x1", u.Meta["x"])
}
