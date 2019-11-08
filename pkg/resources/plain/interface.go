/*
 * Copyright 2019 SAP SE or an SAP affiliate company. All rights reserved.
 * This file is licensed under the Apache Software License, v. 2 except as noted
 * otherwise in the LICENSE file
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 *
 */

package plain

import (
	"github.com/gardener/controller-manager-library/pkg/resources/abstract"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"reflect"
)

type GroupKindProvider = abstract.GroupKindProvider
type ClusterObjectKey = abstract.ClusterObjectKey
type ObjectKey = abstract.ObjectKey
type ObjectMatcher = abstract.ObjectMatcher
type ObjectNameProvider = abstract.ObjectNameProvider
type ObjectName = abstract.ObjectName
type ObjectDataName = abstract.ObjectDataName
type ObjectData = abstract.ObjectData
type Decoder = abstract.Decoder

type ResourcesSource interface {
	Resources() Resources
}

type Object interface {
	abstract.Object
	//runtime.ObjectData
	ResourcesSource

	DeepCopy() Object
	GetResource() Interface
}

type Interface interface {
	GroupKindProvider
	ResourcesSource

	GroupVersionKind() schema.GroupVersionKind

	Wrap(ObjectData) (Object, error)
	New(ObjectName) Object
	IsUnstructured() bool

	ObjectType() reflect.Type
	ListType() reflect.Type
}

type Resources interface {
	ResourcesSource

	ResourceContext() ResourceContext

	Get(spec interface{}) (Interface, error)
	GetByExample(obj runtime.Object) (Interface, error)
	GetByGK(gk schema.GroupKind) (Interface, error)
	GetByGVK(gvk schema.GroupVersionKind) (Interface, error)

	GetUnstructured(spec interface{}) (Interface, error)
	GetUnstructuredByGK(gk schema.GroupKind) (Interface, error)
	GetUnstructuredByGVK(gvk schema.GroupVersionKind) (Interface, error)

	Wrap(obj ObjectData) (Object, error)
	Decode([]byte) (Object, error)
}
