/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package minimal

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type MinimalList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MinimalObject `json:"items"`
}

var _ runtime.Object = &MinimalList{}
var _ metav1.ListInterface = &MinimalList{}

func (o *MinimalList) DeepCopyObject() runtime.Object {
	var n MinimalList
	n.TypeMeta = o.TypeMeta
	o.ListMeta.DeepCopyInto(&n.ListMeta)
	if o.Items != nil {
		in, out := &o.Items, &n.Items
		*out = make([]MinimalObject, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return &n
}
