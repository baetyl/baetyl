package utils

import (
	"fmt"
	"os"
	"sync"
)

// TODO: to improve or remove
type s struct {
	kvs map[string]string
	sync.RWMutex
}

var env = s{kvs: map[string]string{}}

// SetEnv sets env
func SetEnv(key, value string) {
	env.Lock()
	defer env.Unlock()
	env.kvs[key] = value
}

// GetEnv gets env
func GetEnv(key string) string {
	env.RLock()
	defer env.RUnlock()
	return env.kvs[key]
}

// AppendEnv appends envs
func AppendEnv(paramEnv map[string]string, includeHostEnv bool) []string {
	out := []string{}
	if includeHostEnv {
		out = os.Environ()
	}
	for k, v := range paramEnv {
		out = append(out, fmt.Sprintf("%s=%s", k, v))
	}
	env.RLock()
	defer env.RUnlock()
	for k, v := range env.kvs {
		out = append(out, fmt.Sprintf("%s=%s", k, v))
	}
	return out
}
