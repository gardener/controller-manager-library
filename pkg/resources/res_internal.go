/*
 * Copyright 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package resources

import (
	"reflect"

	"github.com/gardener/controller-manager-library/pkg/logger"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Internal interface {
	Interface

	I_objectType() reflect.Type
	I_listType() reflect.Type

	I_create(data ObjectData) (ObjectData, error)
	I_get(data ObjectData) error
	I_update(data ObjectData) (ObjectData, error)
	I_updateStatus(data ObjectData) (ObjectData, error)
	I_delete(data ObjectData) error
}

// _i_resource is the implementation of the internal resource interface used by
// the abstract object.
// To avoid to potentially expose those additional methods the resource
// implementation does NOT implement the internal interface. Instead,
// it uses an internal wrapper object, that implements this interface.
// This wrapper is then passed to the abstract resource implementation
// to be used to implement a broader set of methods in a generic manner.

type _i_resource struct {
	*_resource
}

var _ Internal = &_i_resource{}

func (this *_i_resource) I_objectType() reflect.Type {
	return this.otype
}
func (this *_i_resource) I_listType() reflect.Type {
	return this.ltype
}

func (this *_i_resource) I_update(data ObjectData) (ObjectData, error) {
	logger.Infof("UPDATE %s/%s/%s", this.GroupKind(), data.GetNamespace(), data.GetName())
	result := this.helper.createData()
	return result, this.objectRequest(this.client.Put(), data).
		Body(data).
		Do().
		Into(result)
}

func (this *_i_resource) I_updateStatus(data ObjectData) (ObjectData, error) {
	logger.Infof("UPDATE STATUS %s/%s/%s", this.GroupKind(), data.GetNamespace(), data.GetName())
	result := this.helper.createData()
	return result, this.objectRequest(this.client.Put(), data, "status").
		Body(data).
		Do().
		Into(result)
}

func (this *_i_resource) I_create(data ObjectData) (ObjectData, error) {
	result := this.helper.createData()
	return result, this.resourceRequest(this.client.Post(), data).
		Body(data).
		Do().
		Into(result)
}

func (this *_i_resource) I_get(data ObjectData) error {
	return this.objectRequest(this.client.Get(), data).
		Do().
		Into(data)
}

func (this *_i_resource) I_delete(data ObjectData) error {
	return this.objectRequest(this.client.Delete(), data).
		Body(&metav1.DeleteOptions{}).
		Do().
		Error()
}
