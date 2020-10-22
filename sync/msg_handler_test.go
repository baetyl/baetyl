package sync

import (
	"testing"

	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/stretchr/testify/assert"
)

func TestMsgHandler(t *testing.T) {
	lk := &mockLink{}

	hd := &handler{link: lk}

	msg := &specv1.Message{}
	err := hd.OnMessage(msg)
	assert.NoError(t, err)

	err = hd.OnTimeout()
	assert.NoError(t, err)
}
