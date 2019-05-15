package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	set := NewSet()
	assert.Equal(t, 0, set.length)
	assert.True(t, set.IsEmpty())
	set.Add(1)
	assert.Equal(t, 1, set.length)
	set.Add(2)
	set.Add("Hello")
	assert.True(t, set.Has(2))
	assert.False(t, set.Has(3))
	assert.True(t, set.Has("Hello"))
	assert.False(t, set.Has("World"))
}
