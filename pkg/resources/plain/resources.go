/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package plain

import (
	"reflect"

	"github.com/gardener/controller-manager-library/pkg/resources/abstract"
	"github.com/gardener/controller-manager-library/pkg/resources/errors"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type factory struct {
}

var _ abstract.Factory = factory{}

func (this factory) NewResources(ctx abstract.ResourceContext, factory abstract.Factory) abstract.Resources {
	res := &_resources{}
	res.AbstractResources = abstract.NewAbstractResources(ctx, res, factory)
	return res
}

func (this factory) NewResource(resources abstract.Resources, gvk schema.GroupVersionKind, otype, ltype reflect.Type) (abstract.Resource, error) {
	return resources.(*_resources)._newResource(gvk, otype, ltype)
}

func (this factory) ResolveGVK(ctx abstract.ResourceContext, gk schema.GroupKind, gvks []schema.GroupVersionKind, unrestricted bool) (schema.GroupVersionKind, error) {
	switch len(gvks) {
	case 0:
		return schema.GroupVersionKind{}, errors.ErrUnknownResource.New("group kind", gk)
	case 1:
		return gvks[0], nil
	default:
		return schema.GroupVersionKind{}, errors.New(errors.ERR_NON_UNIQUE_MAPPING, "non unique version mapping for %s", gk)
	}
}

type _resources struct {
	*abstract.AbstractResources
}

var _ Resources = &_resources{}

func adapt(r abstract.Resource, err error) (Interface, error) {
	if r == nil {
		return nil, err
	}
	return r.(Interface), err
}

func (this *_resources) Resources() Resources {
	return this
}

func (this *_resources) ResourceContext() ResourceContext {
	return this.AbstractResources.ResourceContext().(ResourceContext)
}

func (this *_resources) Decode(bytes []byte) (Object, error) {
	data, err := this.AbstractResources.Decode(bytes)
	if err != nil {
		return nil, err
	}
	return this.Wrap(data)
}

func (this *_resources) Get(spec interface{}) (Interface, error) {
	return adapt(this.AbstractResources.Get(spec))
}

func (this *_resources) GetByExample(obj runtime.Object) (Interface, error) {
	return adapt(this.AbstractResources.GetByExample(obj))
}

func (this *_resources) GetByGK(gk schema.GroupKind) (Interface, error) {
	return adapt(this.AbstractResources.GetByGK(gk))
}

func (this *_resources) GetByGVK(gvk schema.GroupVersionKind) (Interface, error) {
	return adapt(this.AbstractResources.GetByGVK(gvk))
}

func (this *_resources) GetUnstructured(spec interface{}) (Interface, error) {
	return adapt(this.AbstractResources.GetUnstructured(spec))
}

func (this *_resources) GetUnstructuredByGK(gk schema.GroupKind) (Interface, error) {
	return adapt(this.AbstractResources.GetUnstructuredByGK(gk))
}

func (this *_resources) GetUnstructuredByGVK(gvk schema.GroupVersionKind) (Interface, error) {
	return adapt(this.AbstractResources.GetUnstructuredByGVK(gvk))
}

func (this *_resources) Wrap(obj ObjectData) (Object, error) {
	h, err := this.GetByExample(obj)
	if err != nil {
		return nil, err
	}
	return h.Wrap(obj)
}

func (this *_resources) _newResource(gvk schema.GroupVersionKind, otype, ltype reflect.Type) (Interface, error) {
	handler := newResource(this.ResourceContext(), otype, ltype, gvk)
	return handler, nil
}
