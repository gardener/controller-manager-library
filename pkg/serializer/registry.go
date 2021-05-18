/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package serializer

import (
	"fmt"
	"reflect"
	"sync"

	"k8s.io/apimachinery/pkg/runtime"
	jsonserializer "k8s.io/apimachinery/pkg/runtime/serializer/json"
)

type key struct {
	// kind (kubernetes resources kind, e.g., infrastructure/machines/generic)
	kind string
	// extensionType (e.g., aws/azure/.../auditlog)
	extensionType string
	// subType (e.g., providerConfig/providerStatus)
	subType string
	// extensionVersion (e.g. aws.cloud.gardener.cloud/v1alpha1)
	extensionVersion string
}

type extensionVersion struct {
	APIVersion string `json:"apiVersion"`
}

var (
	apiVersionField = "APIVersion"
	specField       = "Spec"
	statusField     = "Status"
	typeField       = "Type"

	registry = map[key]reflect.Type{}
	types    = map[reflect.Type]key{}
	lock     sync.Mutex

	caseSensitiveJSONIterator = jsonserializer.CaseSensitiveJSONIterator()
)

func MustRegister(kind, extensionType, subType, extensionVersion string, v interface{}) {
	if err := Register(kind, extensionType, subType, extensionVersion, v); err != nil {
		panic(err)
	}
}

func Register(kind, extensionType, subType, extensionVersion string, v interface{}) error {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return fmt.Errorf("prototype value must finally be a struct")
	}

	if _, ok := t.FieldByName(apiVersionField); !ok {
		return fmt.Errorf("struct must have an '%s' field", apiVersionField)
	}

	lock.Lock()
	defer lock.Unlock()

	k := key{kind, extensionType, subType, extensionVersion}
	registry[k] = t
	types[t] = k

	return nil
}

func CreateObject(kind, extensionType, subType, extensionVersion string) interface{} {
	k := key{kind, extensionType, subType, extensionVersion}

	lock.Lock()
	defer lock.Unlock()

	if t, ok := registry[k]; ok {
		return createElem(t)
	}

	return nil
}

func Marshal(v interface{}) ([]byte, error) {
	if key := getKeyForType(v); key != nil {
		value := reflect.ValueOf(v)
		value.Elem().FieldByName(apiVersionField).SetString(key.extensionVersion)
		return caseSensitiveJSONIterator.Marshal(v)
	}

	return nil, fmt.Errorf("unknown object type %T", v)
}

func MarshalToResource(r runtime.Object, s interface{}) error {
	_, field, err := getVerifiedKeyAndField(r, s)
	if err != nil {
		return err
	}

	data, err := Marshal(s)
	if err != nil {
		return fmt.Errorf("error during marshalling: %+v", err)
	}

	field.Set(reflect.ValueOf(&runtime.RawExtension{
		Raw: data,
	}))

	return nil
}

func Unmarshal(kind, extensionType, subType string, data []byte) (interface{}, error) {
	var version extensionVersion
	if err := caseSensitiveJSONIterator.Unmarshal(data, &version); err != nil {
		return nil, err
	}

	into := CreateObject(kind, extensionType, subType, version.APIVersion)
	if into == nil {
		return nil, fmt.Errorf("not found in registry: (%s, %s, %s, %s)", kind, extensionType, subType, version.APIVersion)
	}

	if err := caseSensitiveJSONIterator.Unmarshal(data, into); err != nil {
		return nil, err
	}
	return into, nil
}

func UnmarshalInto(data []byte, into interface{}) error {
	if key := getKeyForType(into); key == nil {
		return fmt.Errorf("type %T not registered", into)
	}

	return caseSensitiveJSONIterator.Unmarshal(data, into)
}

func UnmarshalFromResource(subType string, r runtime.Object) (interface{}, error) {
	var (
		kind = r.GetObjectKind().GroupVersionKind().Kind
		t    = reflect.ValueOf(r).Elem().FieldByName(specField).FieldByName(typeField).Interface().(string)
	)

	field, err := getSubTypeField(subType, r)
	if err != nil {
		return nil, err
	}

	data, ok := field.Interface().(*runtime.RawExtension)
	if !ok {
		return nil, fmt.Errorf("subType %s is not of type *runtime.RawExtension", subType)
	}

	return Unmarshal(kind, t, subType, data.Raw)
}

func UnmarshalFromResourceInto(r runtime.Object, into interface{}) error {
	key, field, err := getVerifiedKeyAndField(r, into)
	if err != nil {
		return err
	}

	data, ok := field.Interface().(*runtime.RawExtension)
	if !ok {
		return fmt.Errorf("subType %s is not of type *runtime.RawExtension", key.subType)
	}

	return UnmarshalInto(data.Raw, into)
}
