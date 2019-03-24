package utils

import "time"

// Trace print elapsed time
func Trace(name string, log func(format string, args ...interface{})) func() {
	start := time.Now()
	return func() {
		log("%s elapsed time: %v", name, time.Since(start))
	}
}
