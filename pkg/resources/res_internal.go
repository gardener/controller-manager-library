/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"sync"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/gardener/controller-manager-library/pkg/informerfactories"

	"github.com/gardener/controller-manager-library/pkg/logger"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Internal interface {
	Interface
	Resource() Interface

	I_CreateData(name ...ObjectDataName) ObjectData

	I_create(data ObjectData) (ObjectData, error)
	I_get(data ObjectData) error
	I_update(data ObjectData) (ObjectData, error)
	I_updateStatus(data ObjectData) (ObjectData, error)
	I_delete(data ObjectDataName) error

	I_modifyByName(name ObjectDataName, status_only, create bool, modifier Modifier) (Object, bool, error)
	I_modify(data ObjectData, status_only, read, create bool, modifier Modifier) (ObjectData, bool, error)

	I_getInformer(minimal bool, namespace string, optionsFunc TweakListOptionsFunc) (GenericInformer, error)
	I_lookupInformer(minimal bool, namespace string) (GenericInformer, error)
	I_list(namespace string, opts metav1.ListOptions) ([]Object, error)
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
	lock  sync.Mutex
	cache map[bool]GenericInformer
}

var _ Internal = &_i_resource{}

func new_i_resource(r *_resource) *_i_resource {
	return &_i_resource{_resource: r, cache: map[bool]GenericInformer{}}
}

func (this *_i_resource) Resource() Interface {
	return this._resource
}

func (this *_i_resource) I_CreateData(name ...ObjectDataName) ObjectData {
	return this._resource.CreateData(name...)
}

func (this *_i_resource) I_update(data ObjectData) (ObjectData, error) {
	logger.Infof("UPDATE %s/%s/%s", this.GroupKind(), data.GetNamespace(), data.GetName())
	result := this.CreateData()
	return result, this.objectRequest(this.client.Put(), data).
		Body(data).
		Do(context.TODO()).
		Into(result)
}

func (this *_i_resource) I_updateStatus(data ObjectData) (ObjectData, error) {
	logger.Infof("UPDATE STATUS %s/%s/%s", this.GroupKind(), data.GetNamespace(), data.GetName())
	result := this.CreateData()
	return result, this.objectRequest(this.client.Put(), data, "status").
		Body(data).
		Do(context.TODO()).
		Into(result)
}

func (this *_i_resource) I_create(data ObjectData) (ObjectData, error) {
	result := this.CreateData()
	return result, this.resourceRequest(this.client.Post(), data).
		Body(data).
		Do(context.TODO()).
		Into(result)
}

func (this *_i_resource) I_get(data ObjectData) error {
	return this.objectRequest(this.client.Get(), data).
		Do(context.TODO()).
		Into(data)
}

func (this *_i_resource) I_delete(data ObjectDataName) error {
	return this.objectRequest(this.client.Delete(), data).
		Body(&metav1.DeleteOptions{}).
		Do(context.TODO()).
		Error()
}

func (this *_i_resource) I_getInformer(minimal bool, namespace string, optionsFunc TweakListOptionsFunc) (GenericInformer, error) {
	if cached, ok := this.cache[minimal]; ok {
		return cached, nil
	}
	this.lock.Lock()
	defer this.lock.Unlock()

	if cached, ok := this.cache[minimal]; ok {
		return cached, nil
	}

	var informers GenericFilteredInformerFactory
	if minimal {
		informers = this.ResourceContext().SharedInformerFactory().MinimalObject()
	} else if this.IsUnstructured() {
		informers = this.ResourceContext().SharedInformerFactory().Unstructured()
	} else {
		informers = this.ResourceContext().SharedInformerFactory().Structured()
	}
	informer, err := informers.FilteredInformerFor(this.GroupVersionKind(), namespace, optionsFunc)
	if err != nil {
		return nil, err
	}
	if err := informerfactories.Start(this.ResourceContext(), informers, informer.Informer().HasSynced); err != nil {
		return nil, err
	}

	if namespace == "" && optionsFunc == nil {
		this.cache[minimal] = informer
	}
	return informer, nil
}

func (this *_i_resource) I_lookupInformer(minimal bool, namespace string) (GenericInformer, error) {
	if cached, ok := this.cache[minimal]; ok {
		return cached, nil
	}
	this.lock.Lock()
	defer this.lock.Unlock()

	if cached, ok := this.cache[minimal]; ok {
		return cached, nil
	}

	informers := this.ResourceContext().SharedInformerFactory().Structured()
	if this.IsUnstructured() {
		informers = this.ResourceContext().SharedInformerFactory().Unstructured()
	}
	informer, err := informers.LookupInformerFor(this.GroupVersionKind(), namespace)
	if err != nil {
		return nil, err
	}
	if err := informerfactories.Start(this.ResourceContext(), informers, informer.Informer().HasSynced); err != nil {
		return nil, err
	}

	return informer, nil
}

func (this *_i_resource) I_list(namespace string, options metav1.ListOptions) ([]Object, error) {
	result := this.CreateListData()
	err := this.namespacedRequest(this.client.Get(), namespace).VersionedParams(&options, this.GetParameterCodec()).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		return nil, err
	}
	return this.handleList(result)
}

func (this *_i_resource) I_modifyByName(name ObjectDataName, status_only, create bool, modifier Modifier) (Object, bool, error) {
	data := this.CreateData()
	data.SetName(name.GetName())
	data.SetNamespace(name.GetNamespace())

	data, mod, err := this.I_modify(data, status_only, true, create, modifier)
	if err != nil {
		return nil, mod, err
	}
	return this.helper.ObjectAsResource(data), mod, nil
}

func (this *_i_resource) I_modify(data ObjectData, status_only, read, create bool, modifier Modifier) (ObjectData, bool, error) {
	var lasterr error
	var err error

	if read {
		err = this.I_get(data)
	}

	cnt := 10

	if create {
		if err != nil {
			if !errors.IsNotFound(err) {
				return nil, false, err
			}
			_, err := modifier(data)
			if err != nil {
				return nil, false, err
			}
			created, err := this.I_create(data)
			if err == nil {
				return created, true, nil
			}
			if !errors.IsAlreadyExists(err) {
				return nil, false, err
			}
			err = this.I_get(data)
			if err != nil {
				return nil, false, err
			}
		}
	}

	for cnt > 0 {
		mod, err := modifier(data)
		if !mod {
			if err != nil {
				return nil, mod, err
			}
			return data, mod, nil
		}
		if err == nil {
			var modified ObjectData
			if status_only {
				modified, lasterr = this.I_updateStatus(data)
			} else {
				modified, lasterr = this.I_update(data)
			}
			if lasterr == nil {
				return modified, mod, nil
			}
			if !errors.IsConflict(lasterr) {
				return nil, mod, lasterr
			}
			err = this.I_get(data)
		}
		if err != nil {
			return nil, mod, err
		}
		cnt--
	}
	return data, true, lasterr
}
