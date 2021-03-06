/*
SPDX-FileCopyrightText: YEAR SAP SE or an SAP affiliate company and Gardener contributors

SPDX-License-Identifier: Apache-2.0
*/

package crds

import (
	"github.com/gardener/controller-manager-library/pkg/resources/apiextensions"
	"github.com/gardener/controller-manager-library/pkg/utils"
)

var registry = apiextensions.NewRegistry()

func init() {
	var data string
	data = `

---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.2.9
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
  scope: Namespaced
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
            description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
            type: string
          kind:
            description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
            type: string
          metadata:
            type: object
          spec:
            description: ExampleSpec is  the specification for an example object.
            properties:
              data:
                description: Data contains any data stored for this url
                type: object
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
            description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
            type: string
          kind:
            description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
            type: string
          metadata:
            type: object
          spec:
            description: ExampleSpec is  the specification for an example object.
            properties:
              URL:
                description: URL is the address of the example
                type: string
              data:
                description: Data contains any data stored for this url
                type: object
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
	utils.Must(registry.RegisterCRD(data))
}

func AddToRegistry(r apiextensions.Registry) {
	registry.AddToRegistry(r)
}
