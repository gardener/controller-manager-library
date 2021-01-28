/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"reflect"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources/abstract"
	"github.com/gardener/controller-manager-library/pkg/resources/errors"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
)

const ATTR_EVENTSOURCE = "event-source"

type factory struct {
}

var _ abstract.Factory = factory{}

func (this factory) NewResources(ctx abstract.ResourceContext, factory abstract.Factory) abstract.Resources {
	return newResources(ctx.(*resourceContext))
}

func (this factory) NewResource(resources abstract.Resources, gvk schema.GroupVersionKind, otype, ltype reflect.Type) (abstract.Resource, error) {
	return resources.(*_resources).newResource(gvk, otype, ltype)
}

func (this factory) ResolveGVK(ctx abstract.ResourceContext, gk schema.GroupKind, gvks []schema.GroupVersionKind, unrestricted bool) (schema.GroupVersionKind, error) {
	switch len(gvks) {
	case 0:
		if !unrestricted {
			return schema.GroupVersionKind{}, errors.ErrUnknownResource.New("group kind", gk)
		}
		def, err := ctx.(ResourceContext).GetPreferred(gk)
		if err != nil {
			return schema.GroupVersionKind{}, err
		}
		return def.GroupVersionKind(), err
	case 1:
		return gvks[0], nil
	default:
		for _, gvk := range gvks {
			def, err := ctx.(ResourceContext).GetPreferred(gvk.GroupKind())
			if err != nil {
				return schema.GroupVersionKind{}, err
			}
			if def.Version() == gvk.Version {
				return gvk, nil
			}
		}
		return schema.GroupVersionKind{}, errors.New(errors.ERR_NON_UNIQUE_MAPPING, "non unique version mapping for %s", gk)
	}
}

type _resources struct {
	*abstract.AbstractResources
	record.EventRecorder

	informers *sharedInformerFactory
}

var _ Resources = &_resources{}

func adapt(r abstract.Resource, err error) (Interface, error) {
	if r == nil {
		return nil, err
	}
	return r.(Interface), err
}

func newResources(ctx *resourceContext) *_resources {
	source := "controller"
	src := ctx.Value(ATTR_EVENTSOURCE)
	if src != nil {
		source = src.(string)
	}

	res := &_resources{}
	res.AbstractResources = abstract.NewAbstractResources(ctx, res, factory{})

	res.informers = ctx.sharedInformerFactory

	client, _ := ctx.GetClient(schema.GroupVersion{"", "v1"})

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(logger.Debugf)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: typedcorev1.New(client).Events("")})
	res.EventRecorder = eventBroadcaster.NewRecorder(ctx.scheme, corev1.EventSource{Component: source})

	return res
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

func (this *_resources) _get(obj ObjectData) (Interface, error) {
	if u, ok := obj.(*unstructured.Unstructured); ok {
		return this.GetUnstructured(u)
	} else {
		return this.GetByExample(obj)
	}
}

func (this *_resources) Wrap(obj ObjectData) (Object, error) {
	h, err := this._get(obj)
	if err != nil {
		return nil, err
	}
	return h.Wrap(obj)
}

func (this *_resources) CreateObject(obj ObjectData) (Object, error) {
	r, err := this._get(obj)
	if err != nil {
		return nil, err
	}
	return r.Create(obj)
}

func (this *_resources) CreateOrUpdateObject(obj ObjectData) (Object, error) {
	r, err := this._get(obj)
	if err != nil {
		return nil, err
	}
	return r.CreateOrUpdate(obj)
}

func (this *_resources) DeleteObject(obj ObjectData) error {
	r, err := this._get(obj)
	if err != nil {
		return err
	}
	return r.Delete(obj)
}

func (this *_resources) UpdateObject(obj ObjectData) (Object, error) {
	r, err := this._get(obj)
	if err != nil {
		return nil, err
	}
	return r.Update(obj)
}

func (this *_resources) ModifyObject(obj ObjectData, modifier func(data ObjectData) (bool, error)) (ObjectData, bool, error) {
	r, err := this._get(obj)
	if err != nil {
		return nil, false, err
	}
	return r.Modify(obj, modifier)
}

func (this *_resources) ModifyObjectStatus(obj ObjectData, modifier func(data ObjectData) (bool, error)) (ObjectData, bool, error) {
	r, err := this._get(obj)
	if err != nil {
		return nil, false, err
	}
	return r.ModifyStatus(obj, modifier)
}

func (this *_resources) GetObject(spec interface{}) (Object, error) {
	h, err := this.Get(spec)
	if err != nil {
		return nil, err
	}

	return h.Get(spec)
}

func (this *_resources) GetObjectInto1(obj ObjectData) (Object, error) {
	h, err := this.GetByExample(obj)
	if err != nil {
		return nil, err
	}

	return h.GetInto1(obj)
}

func (this *_resources) GetObjectInto(name ObjectName, obj ObjectData) (Object, error) {
	h, err := this.GetByExample(obj)
	if err != nil {
		return nil, err
	}

	return h.GetInto(name, obj)
}

func (this *_resources) GetCachedObject(spec interface{}) (Object, error) {
	h, err := this.Get(spec)
	if err != nil {
		return nil, err
	}

	return h.GetCached(spec)
}

func (this *_resources) newResource(gvk schema.GroupVersionKind, otype, ltype reflect.Type) (Interface, error) {
	return newResource(this.ResourceContext(), otype, ltype, gvk)
}
