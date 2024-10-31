/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

func IsNil(o interface{}) bool {
	if o == nil {
		return true
	}
	v := reflect.ValueOf(o)
	switch v.Kind() {
	case reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func, reflect.Ptr, reflect.UnsafePointer:
		return v.IsNil()
	}
	return false
}

func SetValue(f reflect.Value, v interface{}) error {
	vv := reflect.ValueOf(v)
	if f.Type() != vv.Type() {
		if !vv.Type().ConvertibleTo(f.Type()) {
			return fmt.Errorf("type %s cannot be converted to %s", vv.Type(), f.Type())
		}
		vv = vv.Convert(f.Type())
	}
	if !f.CanSet() {
		if !f.CanInterface() && f.CanAddr() {
			f = reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem() // #nosec G103 -- needed to access unexported fields
		}
	}
	f.Set(vv)
	return nil
}

func GetValue(f reflect.Value) interface{} {
	if !f.CanInterface() && f.CanAddr() {
		f = reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem() // #nosec G103 -- needed to access unexported fields
	}
	return f.Interface()
}

func SplitString(n string, sel func(s string) (string, bool), seps ...string) []string {
	var result []string
	sep := ","
	if len(seps) > 0 {
		sep = seps[0]
	}
	for _, p := range strings.Split(n, sep) {
		if v, ok := sel(p); ok {
			result = append(result, v)
		}
	}
	return result
}

func Sanitize(list []string, sel func(s string) (string, bool)) []string {
	for i := 0; i < len(list); i++ {
		if v, ok := sel(list[i]); ok {
			list[i] = v
		} else {
			list = append(list[:i], list[i+1:]...)
		}
	}
	return list
}

func IsEmptyString(s *string) bool {
	return s == nil || *s == ""
}

func StringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
func Int64Value(v *int64, def int64) int64 {
	if v == nil {
		return def
	}
	return *v
}

func StringEqual(a, b *string) bool {
	return a == b || (a != nil && b != nil && *a == *b)
}
func IntEqual(a, b *int) bool {
	return a == b || (a != nil && b != nil && *a == *b)
}
func Int64Equal(a, b *int64) bool {
	return a == b || (a != nil && b != nil && *a == *b)
}

func Strings(s ...string) string {
	return "[" + strings.Join(s, ", ") + "]"
}

func Interfaces(elems ...interface{}) string {
	r := "["
	sep := ""
	for _, e := range elems {
		r = fmt.Sprintf("%s%s%s", r, sep, e)
		sep = ", "
	}
	return r + "]"
}

func StringArrayAddUnique(array *[]string, values ...string) []string {
values:
	for _, v := range values {
		for _, o := range *array {
			if v == o {
				continue values
			}
		}
		*array = append(*array, v)
	}
	return *array
}
