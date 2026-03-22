package utils

import "reflect"

func NewPointer[T any](t T) *T {
	v := reflect.ValueOf(t)
	switch v.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func, reflect.Interface:
		if v.IsNil() {
			return nil
		} else {
			return &t
		}
	default:
		return &t
	}
}
