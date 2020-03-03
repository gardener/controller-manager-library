// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package example

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
}
