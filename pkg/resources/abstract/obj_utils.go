/*
 * SPDX-FileCopyrightText: Contributors to the Gardener project
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package abstract

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (this *AbstractObject) GetLabel(name string) string {
	labels := this.GetLabels()
	if labels == nil {
		return ""
	}
	return labels[name]
}

func (this *AbstractObject) GetAnnotation(name string) string {
	annos := this.GetAnnotations()
	if annos == nil {
		return ""
	}
	return annos[name]
}

func (this *AbstractObject) GetOwnerReference() *metav1.OwnerReference {
	return metav1.NewControllerRef(this.ObjectData, this.GroupVersionKind())
}

func (this *AbstractObject) IsDeleting() bool {
	return this.GetDeletionTimestamp() != nil
}
