/*
Copyright (c) YEAR SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package crds

import (
	"strings"
)

var CRDS = map[string]map[string]string{}

func add(name, data string) {
	path := strings.Split(name, ".")
	version := path[len(path)-1]
	if version != "v1beta1" {
		version = "v1"
	}

	crds := CRDS[version]
	if crds == nil {
		crds = map[string]string{}
		CRDS[version] = crds
	}
	crds[path[0]] = data
}

func init() {
	var data string
	data = `

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.2.4
  creationTimestamp: null
  name: examples.example.examples.gardener.cloud
spec:
  additionalPrinterColumns:
  - JSONPath: .spec.URL
    name: URL
    type: string
  group: example.examples.gardener.cloud
  names:
    kind: Example
    listKind: ExampleList
    plural: examples
    shortNames:
    - exa
    singular: example
  preserveUnknownFields: false
  scope: namespaced
  validation:
    openAPIV3Schema:
      description: Example is an example for a custom resource.
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
          type: string
        metadata:
          type: object
        spec:
          description: ExampleSpec is  the specification for an example object.
          properties:
            URL:
              description: URL is the address of the example
              type: string
          required:
          - URL
          type: object
      required:
      - spec
      type: object
  version: v1alpha1
  versions:
  - name: v1alpha1
    served: true
    storage: false
  - name: v1beta1
    served: true
    storage: true
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
  `
	add("example.examples.gardener.cloud_examples.v1beta1", data)
	data = `

---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.2.4
  creationTimestamp: null
  name: examples.example.examples.gardener.cloud
spec:
  group: example.examples.gardener.cloud
  names:
    kind: Example
    listKind: ExampleList
    plural: examples
    shortNames:
    - exa
    singular: example
  scope: namespaced
  versions:
  - additionalPrinterColumns:
    - jsonPath: .spec.hostname
      name: Hostname
      type: string
    - jsonPath: .spec.scheme
      name: URLScheme
      type: string
    - jsonPath: .spec.path
      name: Path
      type: string
    - jsonPath: .spec.port
      name: Port
      type: number
    name: v1alpha1
    schema:
      openAPIV3Schema:
        description: Example is an example for a custom resource.
        properties:
          apiVersion:
            description: 'APIVersion defines the versioned schema of this representation
              of an object. Servers should convert recognized schemas to the latest
              internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
            type: string
          kind:
            description: 'Kind is a string value representing the REST resource this
              object represents. Servers may infer this from the endpoint the client
              submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
            type: string
          metadata:
            type: object
          spec:
            description: ExampleSpec is  the specification for an example object.
            properties:
              hostname:
                description: Hostname is a host name
                type: string
              path:
                description: Path is a path for the URL
                type: string
              port:
                description: Port is a port name for the URL
                type: integer
              scheme:
                description: URLScheme is an URL scheme name to compose an url
                type: string
            required:
            - hostname
            - scheme
            type: object
        required:
        - spec
        type: object
    served: true
    storage: false
    subresources: {}
  - additionalPrinterColumns:
    - jsonPath: .spec.URL
      name: URL
      type: string
    name: v1beta1
    schema:
      openAPIV3Schema:
        description: Example is an example for a custom resource.
        properties:
          apiVersion:
            description: 'APIVersion defines the versioned schema of this representation
              of an object. Servers should convert recognized schemas to the latest
              internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
            type: string
          kind:
            description: 'Kind is a string value representing the REST resource this
              object represents. Servers may infer this from the endpoint the client
              submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
            type: string
          metadata:
            type: object
          spec:
            description: ExampleSpec is  the specification for an example object.
            properties:
              URL:
                description: URL is the address of the example
                type: string
            required:
            - URL
            type: object
        required:
        - spec
        type: object
    served: true
    storage: true
    subresources: {}
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
  `
	add("example.examples.gardener.cloud_examples", data)
	data = `

  `
	add("manifests.go", data)
}
