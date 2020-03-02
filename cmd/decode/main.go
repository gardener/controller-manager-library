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

package main

import (
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/examples/apis/example"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/examples/apis/example/install"
	"github.com/gardener/controller-manager-library/pkg/resources"
)

var scheme *runtime.Scheme
var decoder *resources.Decoder

func init() {
	scheme = runtime.NewScheme()
	utilruntime.Must(install.AddToScheme(scheme))
	decoder = resources.NewDecoder(scheme)
}

var manifest = `
apiVersion: example.examples.gardener.cloud/v1alpha1
kind: Example
metadata:
  annotations:
    a.b: o
  name: my.example
  namespace: default
spec:
  hostname: myhost
  path: a/b
  scheme: http
`

func main() {

	o, gvk, err := decoder.Decode([]byte(manifest))
	if err != nil {
		fmt.Printf("err: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("GVK: %s, type: %T\n", gvk, o)

	into := &example.Example{}
	err = scheme.Convert(o, into, nil)
	if err != nil {
		fmt.Printf("convert err: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("url: %s\n", into.Spec.URL)

	into = &example.Example{}
	//beta:=&v1beta1.Example{}
	err = decoder.DecodeInto([]byte(manifest), into)
	if err != nil {
		fmt.Printf("err: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("url: %s\n", into.Spec.URL)

	vers := &runtime.VersionedObjects{}
	err = decoder.DecodeInto([]byte(manifest), vers)
	if err != nil {
		fmt.Printf("vers err: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("len %d: wire: %T, final: %T: %s\n", len(vers.Objects), vers.First(), vers.Last(), vers.Last().(*example.Example).Spec.URL)
}
