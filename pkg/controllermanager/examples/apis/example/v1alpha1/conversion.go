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
 */

package v1alpha1

import (
	"fmt"
	"net/url"
	"strconv"

	"k8s.io/apimachinery/pkg/conversion"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/examples/apis/example"
)

func Convert_v1alpha1_ExampleSpec_To_example_ExampleSpec(in *ExampleSpec, out *example.ExampleSpec, s conversion.Scope) error {
	if in.Port > 0 {
		out.URL = fmt.Sprintf("%s://%s:%d/%s", in.URLScheme, in.Hostname, in.Port, in.Path)
	} else {
		out.URL = fmt.Sprintf("%s://%s/%s", in.URLScheme, in.Hostname, in.Path)
	}
	return nil
}

func Convert_example_ExampleSpec_To_v1alpha1_ExampleSpec(in *example.ExampleSpec, out *ExampleSpec, s conversion.Scope) error {
	u, err := url.Parse(in.URL)
	if err == nil {
		out.URLScheme = u.Scheme
		out.Hostname = u.Hostname()
		out.Path = u.Path
		out.Port, err = strconv.Atoi(u.Port())
		if err != nil {
			out.Port = 0
			err = nil
		}
	}
	return err
}
