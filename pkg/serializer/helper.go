/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package serializer

import (
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/runtime"
)

func createElem(t reflect.Type) interface{} {
	return reflect.New(t).Interface()
}

func getKeyForType(v interface{}) *key {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if k, ok := types[t]; ok {
		return &k
	}
	return nil
}

func getSubTypeField(subType string, r runtime.Object) (reflect.Value, error) {
	var (
		value  = reflect.ValueOf(r)
		spec   = value.Elem().FieldByName(specField)
		status = value.Elem().FieldByName(statusField)
		field  = spec.FieldByName(subType)
	)

	if !field.IsValid() {
		field = status.FieldByName(subType)
	}
	if !field.IsValid() {
		return reflect.Value{}, fmt.Errorf("subType %s not found as field in object", subType)
	}

	return field, nil
}

func getVerifiedKeyAndField(r runtime.Object, s interface{}) (*key, reflect.Value, error) {
	var (
		kind = r.GetObjectKind().GroupVersionKind().Kind

		value = reflect.ValueOf(r)
		spec  = value.Elem().FieldByName(specField)
		t     = spec.FieldByName(typeField).Interface().(string)
	)

	key := getKeyForType(s)
	if key == nil {
		return nil, reflect.Value{}, fmt.Errorf("unknown structure type %T", s)
	}
	if key.kind != kind {
		return nil, reflect.Value{}, fmt.Errorf("resources kind (%s) does not match registered kind (%s) for given object type", kind, key.kind)
	}
	if key.extensionType != t {
		return nil, reflect.Value{}, fmt.Errorf("resources extensionType type (%s) does not match registered type (%s) for given object type", t, key.extensionType)
	}

	field, err := getSubTypeField(key.subType, r)
	if err != nil {
		return nil, reflect.Value{}, err
	}

	if _, ok := field.Interface().(*runtime.RawExtension); !ok {
		return nil, reflect.Value{}, fmt.Errorf("subType %s is not of type *runtime.RawExtension", key.subType)
	}

	return key, field, nil
}
