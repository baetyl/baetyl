package store

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewBoltHold(t *testing.T) {
	f, err := os.CreateTemp("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())

	s1, err := NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, s1)

	type ts struct {
		I int
		S string
	}

	input := &ts{I: 1, S: "s"}
	err = s1.Insert("1", input)
	assert.NoError(t, err)
	input.I = 2
	err = s1.Insert("1", input)
	assert.EqualError(t, err, "This Key already exists in this bolthold for this type")
	err = s1.Upsert("1", input)
	assert.NoError(t, err)

	var output ts
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s2, err := NewBoltHold(f.Name())
		assert.NoError(t, err)
		assert.NotNil(t, s2)

		err = s2.Get("1", &output)
		assert.NoError(t, err)
	}()

	time.Sleep(time.Millisecond * 100)
	s1.Close()
	wg.Wait()

	assert.Equal(t, 2, output.I)
	assert.Equal(t, "s", output.S)
}
