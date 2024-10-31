/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package types

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/gardener/controller-manager-library/pkg/convert"
	"github.com/gardener/controller-manager-library/pkg/types/infodata/simple"
)

// Values is a workarround for kubebuilder to be able to generate
// an API spec. The Values MUST be marked with "-" to avoud errors.

// Values is used to specify an arbitrary document structure
// without the need of a regular manifest api group version
// as part of a kubernetes resource
type Values struct {
	simple.Values `json:"-"`
}

func (in *Values) DeepCopy() *Values {
	if in == nil {
		return nil
	}
	return &Values{in.Values.DeepCopy()}
}

func (this Values) MarshalJSON() ([]byte, error) {
	if this.Values == nil {
		return []byte("null"), nil
	}
	return this.Values.Marshal()
}

func (this *Values) UnmarshalJSON(in []byte) error {
	if this == nil {
		return errors.New("Values: UnmarshalJSON on nil pointer")
	}
	if !bytes.Equal(in, []byte("null")) {
		return json.Unmarshal(in, &this.Values)
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Values) DeepCopyInto(out *Values) {
	clone := in.DeepCopy()
	*out = *clone
}

////////////////////////////////////////////////////////////////////////////////

var mapType = reflect.TypeOf(map[string]interface{}{})
var listType = reflect.TypeOf([]interface{}{})
var stringType = convert.StringType()

func MapType() reflect.Type {
	return mapType
}

func ListType() reflect.Type {
	return listType
}

func NormValues(v simple.Values) simple.Values {
	return simple.Values(CopyAndNormalize(v).(map[string]interface{}))
}

func CopyAndNormalize(in interface{}) interface{} {
	if in == nil {
		return in
	}
	switch e := in.(type) {
	case map[string]string:
		r := map[string]interface{}{}
		for k, v := range e {
			r[k] = CopyAndNormalize(v)
		}
		return r
	case map[string]interface{}:
		r := map[string]interface{}{}
		for k, v := range e {
			r[k] = CopyAndNormalize(v)
		}
		return r
	case Values:
		r := map[string]interface{}{}
		for k, v := range e.Values {
			r[k] = CopyAndNormalize(v)
		}
		return r
	case simple.Values:
		r := map[string]interface{}{}
		for k, v := range e {
			r[k] = CopyAndNormalize(v)
		}
		return r
	case []interface{}:
		r := []interface{}{}
		for _, v := range e {
			r = append(r, CopyAndNormalize(v))
		}
		return r
	case []string:
		r := []interface{}{}
		for _, v := range e {
			r = append(r, v)
		}
		return r
	case string, bool:
		return e
	case int:
		return int64(e)
	case int8:
		return int64(e)
	case int16:
		return int64(e)
	case int32:
		return int64(e)
	case uint:
		return int64(e)
	case uint8:
		return int64(e)
	case uint16:
		return int64(e)
	case uint32:
		return int64(e)
	case uint64:
		return int64(e)
	case float32:
		return float64(e)

	case int64, float64:
		return e
	default:
		value := reflect.ValueOf(e)
		t := value.Type()
		switch t.Kind() {
		case reflect.Map:
			if t.ConvertibleTo(mapType) {
				return CopyAndNormalize(value.Convert(mapType).Interface())
			}
			if t.Key().ConvertibleTo(stringType) {
				r := map[string]interface{}{}
				iter := value.MapRange()
				for iter.Next() {
					k := iter.Key()
					v := iter.Value()
					r[k.Convert(stringType).Interface().(string)] = CopyAndNormalize(v.Interface())
				}
				return r
			}
			r, err := convert.ConvertTo(value, mapType)
			if err == nil {
				return CopyAndNormalize(r)
			}
		case reflect.Array, reflect.Slice:
			if t.ConvertibleTo(listType) {
				return CopyAndNormalize(value.Convert(listType).Interface())
			}
			r := make([]interface{}, value.Len())
			for i := 0; i < value.Len(); i++ {
				r[i] = CopyAndNormalize(value.Index(i).Interface())
			}
			return r
		default:
			switch t.Kind() {
			case reflect.String:
				return value.Convert(convert.StringType()).Interface()
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				return value.Convert(convert.Int64Type()).Interface()
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				return value.Convert(convert.Int64Type()).Interface()
			case reflect.Float32, reflect.Float64:
				return value.Convert(convert.Float64Type()).Interface()
			case reflect.Bool:
				return value.Convert(convert.BoolType()).Interface()
			}
		}
		panic(fmt.Errorf("invalid type %T", e))
	}
}