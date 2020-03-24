package event

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/baetyl/baetyl-core/store"
	"github.com/stretchr/testify/assert"
)

func TestCenter(t *testing.T) {
	f, err := ioutil.TempFile("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())

	s, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, s)

	c, err := NewCenter(s, 2)
	assert.NoError(t, err)
	assert.NotNil(t, c)

	events := make(chan *Event, 2)
	err = c.Register(t.Name(), func(e *Event) error {
		fmt.Println("-->2 handling", e.String())
		events <- e
		return nil
	})
	assert.NoError(t, err)
	c.Start()

	go func() {
		e1 := NewEvent(t.Name(), []byte("test"))
		err = c.Trigger(e1)
		assert.NoError(t, err)
		err = c.Trigger(e1)
		assert.NoError(t, err)
		err = c.Trigger(e1)
		assert.NoError(t, err)
		err = c.Trigger(e1)
		assert.NoError(t, err)
	}()

	e2 := <-events
	assert.Equal(t, uint64(1), e2.ID)
	e2 = <-events
	assert.Equal(t, uint64(2), e2.ID)
	e2 = <-events
	assert.Equal(t, uint64(3), e2.ID)
	e2 = <-events
	assert.Equal(t, uint64(4), e2.ID)
	assert.Equal(t, []byte("test"), e2.Payload)
}

func TestCenterException(t *testing.T) {
	f, err := ioutil.TempFile("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())

	s, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, s)

	handler := func(e *Event) error {
		fmt.Println("-->handling", e.String())
		return os.ErrInvalid
	}

	c, err := NewCenter(nil, 2)
	assert.Equal(t, os.ErrInvalid, err)
	assert.Nil(t, c)

	c, err = NewCenter(s, 0)
	assert.Equal(t, os.ErrInvalid, err)
	assert.Nil(t, c)

	c, err = NewCenter(s, 2)
	assert.NoError(t, err)
	assert.NotNil(t, c)

	err = c.Register("", handler)
	assert.Equal(t, os.ErrInvalid, err)

	err = c.Register("2", nil)
	assert.Equal(t, os.ErrInvalid, err)

	err = c.Register("2", handler)
	assert.NoError(t, err)

	var e1 *Event
	err = c.Trigger(e1)
	assert.NoError(t, err)

	e1 = NewEvent("", nil)
	err = c.Trigger(e1)
	assert.Equal(t, os.ErrInvalid, err)
}
