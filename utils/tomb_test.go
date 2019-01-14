package utils_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/baidu/openedge/module/utils"
	"github.com/stretchr/testify/assert"
)

func TestTomb(t *testing.T) {
	tb := new(utils.Tomb)
	err := tb.Go(func() error {
		<-tb.Dying()
		return nil
	})
	assert.NoError(t, err)
	tb.Kill(nil)
	err = tb.Wait()
	assert.NoError(t, err)

	tb = new(utils.Tomb)
	err = tb.Go(func() error {
		<-tb.Dying()
		return fmt.Errorf("abc")
	})
	assert.NoError(t, err)
	tb.Kill(nil)
	tb.Kill(nil)
	err = tb.Wait()
	assert.EqualError(t, err, "abc")
	err = tb.Wait()
	assert.EqualError(t, err, "abc")

	tb = new(utils.Tomb)
	tb.Kill(fmt.Errorf("abc"))
	err = tb.Go(func() error {
		<-tb.Dying()
		return nil
	})
	assert.NoError(t, err)
	err = tb.Wait()
	assert.EqualError(t, err, "abc")

	tb = new(utils.Tomb)
	tb.Kill(fmt.Errorf("abc"))
	err = tb.Wait()
	assert.NoError(t, err)
	err = tb.Go(func() error {
		<-tb.Dying()
		return nil
	})
	assert.NoError(t, err)
	tb.Kill(nil)
	err = tb.Wait()
	assert.EqualError(t, err, "abc")

	tb = new(utils.Tomb)
	tb.Kill(nil)
	err = tb.Wait()
	assert.NoError(t, err)
	err = tb.Go(func() error {
		<-tb.Dying()
		return nil
	})
	assert.NoError(t, err)
	tb.Kill(fmt.Errorf("abc"))
	err = tb.Wait()
	assert.EqualError(t, err, "abc")

	tb = new(utils.Tomb)
	err = tb.Go(func() error {
		<-tb.Dying()
		return nil
	})
	assert.NoError(t, err)
	tb.Kill(nil)
	err = tb.Wait()
	assert.NoError(t, err)
	err = tb.Go(func() error {
		<-tb.Dying()
		return nil
	})
	assert.EqualError(t, err, "tomb.Go called after all goroutines terminated")

	tb = new(utils.Tomb)
	err = tb.Go(func() error {
		return nil
	})
	assert.NoError(t, err)
	time.Sleep(time.Millisecond * 100)
	err = tb.Go(func() error {
		return nil
	})
	assert.EqualError(t, err, "tomb.Go called after all goroutines terminated")
	tb.Kill(nil)
	err = tb.Wait()
	assert.NoError(t, err)
}

func BenchmarkA(b *testing.B) {
	msg := "aaa"
	msgchan := make(chan string, b.N)
	var tomb utils.Tomb
	for i := 0; i < b.N; i++ {
		select {
		case <-tomb.Dying():
			continue
		case msgchan <- msg:
		default: // discard if channel is full
		}
	}
}

func BenchmarkB(b *testing.B) {
	msg := "aaa"
	msgchan := make(chan string, b.N)
	var tomb utils.Tomb
	for i := 0; i < b.N; i++ {
		if !tomb.Alive() {
			continue
		}
		select {
		case msgchan <- msg:
		default: // discard if channel is full
		}
	}
}
