package router

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeque(t *testing.T) {
	d := newDeque()

	d.Push(1)
	d.Push("abc")
	d.Push([]string{"hello", "world"})

	assert.Equal(t, 3, d.Len())
	assert.Equal(t, []string{"hello", "world"}, d.Peek())
	assert.Equal(t, []string{"hello", "world"}, d.Pop())
	assert.Equal(t, 2, d.Len())

	d.Offer([]string{"hello", "world"})
	assert.Equal(t, 1, d.Poll())
	assert.Equal(t, 2, d.Len())
	assert.Equal(t, false, d.Empty())

	d.Pop()
	d.Pop()
	assert.Equal(t, true, d.Empty())
}
