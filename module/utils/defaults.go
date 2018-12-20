package utils

import (
	"fmt"
	"reflect"

	"github.com/creasty/defaults"
)

// SetDefaults set default values
func SetDefaults(ptr interface{}) error {
	err := defaults.Set(ptr)
	if err != nil {
		return fmt.Errorf("%v: %s", ptr, err.Error())
	}

	v := reflect.ValueOf(ptr).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		if f := t.Field(i); f.Type.Kind() == reflect.Slice {
			slice := v.Field(i)
			for j := 0; j < slice.Len(); j++ {
				sliceItem := slice.Index(j)
				if sliceItem.Kind() != reflect.Struct {
					continue
				}
				sliceItemTemp := reflect.New(sliceItem.Type())
				sliceItemTemp.Elem().Set(sliceItem)
				err = SetDefaults(sliceItemTemp.Interface())
				if err != nil {
					return err
				}
				sliceItem.Set(sliceItemTemp.Elem())
			}
		}
	}
	return nil
}
