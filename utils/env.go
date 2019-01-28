package utils

import (
	"fmt"
	"os"
	"sync"
)

type s struct {
	kvs map[string]string
	sync.Mutex
}

var env = s{kvs: map[string]string{}}

// SetEnv sets env
func SetEnv(key, value string) error {
	env.Lock()
	defer env.Unlock()
	env.kvs[key] = value
	return os.Setenv(key, value)
}

// GetEnv gets env
func GetEnv(key string) string {
	return os.Getenv(key)
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
	env.Lock()
	defer env.Unlock()
	for k, v := range env.kvs {
		out = append(out, fmt.Sprintf("%s=%s", k, v))
	}
	return out
}
