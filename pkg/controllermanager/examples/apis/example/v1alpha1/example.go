// SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ExampleList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Example `json:"items"`
}

// +kubebuilder:printcolumn:name=Hostname,JSONPath=".spec.hostname",type=string
// +kubebuilder:printcolumn:name=URLScheme,JSONPath=".spec.scheme",type=string
// +kubebuilder:printcolumn:name=Path,JSONPath=".spec.path",type=string
// +kubebuilder:printcolumn:name=Port,JSONPath=".spec.port",type=number
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Example is an example for a custom resource.
type Example struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ExampleSpec `json:"spec"`
}

// ExampleSpec is  the specification for an example object.
type ExampleSpec struct {
	// Hostname is a host name
	Hostname string `json:"hostname"`
	// URLScheme is an URL scheme name to compose an url
	URLScheme string `json:"scheme"`
	// Port is a port name for the URL
	// +optional
	Port int `json:"port,omitempty"`
	// Path is a path for the URL
	// +optional
	Path string `json:"path,omitempty"`
	// Data contains any data stored for this url
	// +optional
	Data *runtime.RawExtension `json:"data,omitempty"`
}
