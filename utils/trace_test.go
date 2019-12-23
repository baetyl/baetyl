package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Trace print elapsed time
func TestTrace(t *testing.T) {
	f := func(format string, args ...interface{}) {}
	trace := Trace("end to clean,", f)
	assert.NotEmpty(t, trace)
}
