package utils

import (
	"reflect"
)

// Equal compares two struct data
func Equal(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}
