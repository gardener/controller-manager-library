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
	"encoding/json"
	"fmt"
	"os"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/webhook/conversion/api"
	v1 "github.com/gardener/controller-manager-library/pkg/controllermanager/webhook/conversion/api/v1"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/webhook/conversion/api/v1beta1"
	"github.com/gardener/controller-manager-library/pkg/resources"
)

var scheme *runtime.Scheme
var decoder *resources.Decoder

func init() {
	scheme = runtime.NewScheme()
	utilruntime.Must(v1.AddToScheme(scheme))
	utilruntime.Must(v1beta1.AddToScheme(scheme))
	utilruntime.Must(api.AddToScheme(scheme))
	decoder = resources.NewDecoder(scheme)
}

var manifest = `
apiVersion: apiextensions.k8s.io/v1
kind: ConversionReview
request:
  uid: bla-blub
  objects:
  - apiVersion: example.examples.gardener.cloud/v1alpha1
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

	versions := &runtime.VersionedObjects{}

	if err := decoder.DecodeInto([]byte(manifest), versions); err != nil {
		fmt.Printf("err: %s\n", err)
		os.Exit(1)
	}

	obj := versions.First()
	req := versions.Last().(*api.ConversionReview).Request

	fmt.Printf("GK: %s, T: %T\n", obj.GetObjectKind().GroupVersionKind(), obj)
	resp := &api.ConversionResponse{
		UID:              req.UID,
		ConvertedObjects: req.Objects,
		Result:           meta.Status{},
	}

	answer := &api.ConversionReview{
		Response: resp,
	}
	answer.SetGroupVersionKind(obj.GetObjectKind().GroupVersionKind())

	err := scheme.Convert(answer, obj, nil)
	if err != nil {
		fmt.Printf("err: %s\n", err)
		os.Exit(1)
	}

	out, err := json.MarshalIndent(answer, "", "  ")
	if err != nil {
		fmt.Printf("fmt err: 5s\n", err)
		os.Exit(1)
	}
	fmt.Printf("%s\n", string(out))
}
