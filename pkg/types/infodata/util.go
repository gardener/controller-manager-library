/*
 * Copyright 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
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
	if t.Kind() == reflect.Ptr {
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
