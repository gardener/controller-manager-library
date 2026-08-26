/*
 * SPDX-FileCopyrightText: Copyright Contributors to the Gardener project
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 *
 */

package infodata

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func UnmarshalFunc(elem InfoData) Unmarshaller {
	t := reflect.TypeOf(elem)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return func(bytes []byte) (InfoData, error) {
		if bytes == nil {
			return nil, fmt.Errorf("no data given")
		}
		data := reflect.New(t)
		err := json.Unmarshal(bytes, data.Interface())
		if err != nil {
			return nil, err
		}
		return data.Elem().Interface().(InfoData), nil
	}
}
