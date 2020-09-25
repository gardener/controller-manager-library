/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package main

import (
	"fmt"
	"os"

	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/examples/apis/example"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/examples/apis/example/v1alpha1"
	"github.com/gardener/controller-manager-library/pkg/resources"
)

var scheme *runtime.Scheme
var decoder *resources.Decoder

func init() {
	scheme = runtime.NewScheme()
	utilruntime.Must(example.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
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

	vers := &resources.VersionedObjects{}
	err = decoder.DecodeInto([]byte(manifest), vers)
	if err != nil {
		fmt.Printf("vers err: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("len %d: wire: %T, final: %T: %s\n", len(vers.Objects), vers.First(), vers.Last(), vers.Last().(*example.Example).Spec.URL)

	txt := `
description: "ACMEIssuerDNS01ProviderAkamai is a structure containing
	the DNS configuration for Akamai DNS&#0978;Zone Record Management
	API"
`

	m := map[string]interface{}{}
	err = yaml.Unmarshal([]byte(txt), &m)
	if err != nil {
		fmt.Printf("Unmarshal: %s\n", err)
	} else {
		fmt.Printf("value: %s\n", m["description"])
	}
}
