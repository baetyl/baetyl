package event

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"
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

func TestCenterRenew(t *testing.T) {
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

	// insert some events before center starts
	e1 := NewEvent(t.Name(), []byte("t1"))
	err = c.Trigger(e1)
	assert.NoError(t, err)
	err = c.Trigger(e1)
	assert.NoError(t, err)
	c.Close()

	// new center again
	c, err = NewCenter(s, 2)
	assert.NoError(t, err)
	assert.NotNil(t, c)
	assert.Equal(t, uint64(2), c.last)

	events := make(chan *Event, 4)
	err = c.Register(t.Name(), func(e *Event) error {
		fmt.Println("-->2 handling", e.String())
		events <- e
		return nil
	})
	assert.NoError(t, err)
	c.Start()

	go func() {
		e2 := NewEvent(t.Name(), []byte("t2"))
		err = c.Trigger(e2)
		assert.NoError(t, err)
		err = c.Trigger(e2)
		assert.NoError(t, err)
	}()

	e3 := <-events
	assert.Equal(t, uint64(1), e3.ID)
	assert.Equal(t, []byte("t1"), e3.Payload)
	e3 = <-events
	assert.Equal(t, uint64(2), e3.ID)
	assert.Equal(t, []byte("t1"), e3.Payload)
	e3 = <-events
	assert.Equal(t, uint64(3), e3.ID)
	assert.Equal(t, []byte("t2"), e3.Payload)
	e3 = <-events
	assert.Equal(t, uint64(4), e3.ID)
	assert.Equal(t, []byte("t2"), e3.Payload)
}

func TestCenterException(t *testing.T) {
	f, err := ioutil.TempFile("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())

	s, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, s)

	c, err := NewCenter(nil, 2)
	assert.Equal(t, os.ErrInvalid, err)
	assert.Nil(t, c)

	c, err = NewCenter(s, 0)
	assert.Equal(t, os.ErrInvalid, err)
	assert.Nil(t, c)

	c, err = NewCenter(s, 2)
	assert.NoError(t, err)
	assert.NotNil(t, c)
	defer c.Close()
	c.Start()

	var wg sync.WaitGroup
	wg.Add(1)
	handler := func(e *Event) error {
		defer wg.Done()
		fmt.Println("-->handling", e.String())
		return os.ErrInvalid
	}

	err = c.Register("", handler)
	assert.Equal(t, os.ErrInvalid, err)

	err = c.Register("2", nil)
	assert.Equal(t, os.ErrInvalid, err)

	err = c.Register("handler", handler)
	assert.NoError(t, err)

	var e1 *Event
	err = c.Trigger(e1)
	assert.NoError(t, err)

	e1 = NewEvent("", nil)
	err = c.Trigger(e1)
	assert.Equal(t, os.ErrInvalid, err)

	e1 = NewEvent("no-handler", nil)
	err = c.Trigger(e1)
	assert.NoError(t, err)

	e1 = NewEvent("handler", nil)
	err = c.Trigger(e1)
	assert.NoError(t, err)
	// wait until event is handled
	wg.Wait()

	c.Close()
	// persist an event which will be handled when center restarts
	err = c.Trigger(e1)
	assert.NoError(t, err)

	// close the store to simulate errors from store
	s.Close()

	err = c.Trigger(e1)
	assert.EqualError(t, err, "database not open")

	c, err = NewCenter(s, 2)
	assert.EqualError(t, err, "database not open")
	assert.Nil(t, c)
}
