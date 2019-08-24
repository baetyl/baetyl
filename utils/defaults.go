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
		tf := t.Field(i)
		vf := v.Field(i)
		if tf.Type.Kind() == reflect.Slice {
			for j := 0; j < vf.Len(); j++ {
				item := vf.Index(j)
				if item.Kind() != reflect.Struct {
					continue
				}
				err := setDefaults(item)
				if err != nil {
					return err
				}
			}
		}
		if tf.Type.Kind() == reflect.Map {
			for _, k := range vf.MapKeys() {
				item := vf.MapIndex(k)
				if item.Kind() != reflect.Struct {
					continue
				}
				tmp := reflect.New(item.Type())
				tmp.Elem().Set(item)
				err := setDefaults(tmp.Elem())
				vf.SetMapIndex(k, tmp.Elem())
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func setDefaults(v reflect.Value) error {
	tmp := reflect.New(v.Type())
	tmp.Elem().Set(v)
	err := SetDefaults(tmp.Interface())
	if err != nil {
		return err
	}
	v.Set(tmp.Elem())
	return nil
}
