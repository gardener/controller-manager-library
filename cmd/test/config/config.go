/*
 * Copyright 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 *
 */

package config

import (
	"fmt"
	"reflect"
)

type Mine struct {
	Option string `config:"option,'dies ist ein test'"`
}


func ConfigMain() {
	evaluate(&Mine{})
}

func evaluate(o interface{}) {
	t:=reflect.TypeOf(o)

	if t.Kind()==reflect.Ptr {
		t=t.Elem()
	}

	if t.Kind()!=reflect.Struct {
		fmt.Printf("No struct\n")
		return
	}

	for i:= 0; i<t.NumField(); i++ {
		f:=t.Field(i)
		f.Tag.Lookup("config")
		fmt.Printf("%s: %s:  %s\n", f.Name, f.Type, f.Tag.Get("config") )
	}
}