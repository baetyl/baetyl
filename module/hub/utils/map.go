package utils

import (
	"fmt"
	"reflect"
	"strings"
)

// GetKeys gets all keys of map
func GetKeys(m map[string]struct{}) []string {
	keys := reflect.ValueOf(m).MapKeys()
	result := make([]string, 0)
	for _, key := range keys {
		result = append(result, key.Interface().(string))
	}
	return result
}

// Append appends map
func Append(a []string, m map[string]string) []string {
	for k, v := range m {
		a = append(a, KV2S(k, v))
	}
	return a
}

// KV2S generates string from key-value
func KV2S(k string, v interface{}) string {
	return fmt.Sprintf("%s=%v", k, v)
}

// M2A generates array from map
func M2A(m map[string]interface{}) []string {
	a := make([]string, 0)
	for k, v := range m {
		a = append(a, KV2S(k, v))
	}
	return a
}

// A2S generates string from array
func A2S(a []string) string {
	return strings.Join(a, " ")
}

// M2S generates string from map
func M2S(m map[string]interface{}) string {
	return A2S(M2A(m))
}
