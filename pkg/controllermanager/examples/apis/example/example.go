// SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package example

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ExampleList struct {
	metav1.TypeMeta
	// Standard list metadata
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	metav1.ListMeta
	Items []Example
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Example is an example for a custom resource.
type Example struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec ExampleSpec
}

// ExampleSpec is  the specification for an example object.
type ExampleSpec struct {
	URL string

	/*
		// Hostname is a host name
		Hostname string
		// URLScheme is an URL scheme name to compose an url
		URLScheme string
		// Port is a port name for the URL
		Port int
		// Path is a path for the URL
		Path string
	*/

	// Data contains any data stored for this url
	Data *runtime.RawExtension
}
