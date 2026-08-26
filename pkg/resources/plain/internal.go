/*
 * SPDX-FileCopyrightText: Contributors to the Gardener project
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package plain

import (
	"k8s.io/apimachinery/pkg/runtime"
)

type Internal interface {
	Interface

	CreateData(name ...ObjectDataName) ObjectData
	CreateListData() runtime.Object
	CheckOType(obj ObjectData, unstructured ...bool) error
}
