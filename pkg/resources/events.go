/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"k8s.io/client-go/tools/cache"

	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources/minimal"
)

func convert(resource Interface, funcs *ResourceEventHandlerFuncs) *cache.ResourceEventHandlerFuncs {
	return &cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if funcs.AddFunc == nil {
				return
			}
			o, err := resource.Wrap(obj.(ObjectData))
			if err == nil {
				funcs.AddFunc(o)
			}
		},
		DeleteFunc: func(obj interface{}) {
			if funcs.DeleteFunc == nil {
				return
			}
			data, ok := obj.(ObjectData)
			if !ok {
				stale, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					logger.Errorf("informer %q reported unknown object to be deleted (%T)", resource.Name(), obj)
					return
				}
				if stale.Obj == nil {
					logger.Errorf("informer %q reported no stale object to be deleted", resource.Name())
					return
				}
				data, ok = stale.Obj.(ObjectData)
				if !ok {
					logger.Errorf("informer %q reported unknown stale object to be deleted (%T)", resource.Name(), stale.Obj)
					return
				}
			}
			o, err := resource.Wrap(data)
			if err == nil {
				funcs.DeleteFunc(o)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			if funcs.UpdateFunc == nil {
				return
			}
			o, err := resource.Wrap(old.(ObjectData))
			if err == nil {
				n, err := resource.Wrap(new.(ObjectData))
				if err == nil {
					funcs.UpdateFunc(o, n)
				}
			}
		},
	}
}

func convertInfo(resource Interface, funcs *ResourceInfoEventHandlerFuncs) *cache.ResourceEventHandlerFuncs {
	return &cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if funcs.AddFunc == nil {
				return
			}
			funcs.AddFunc(wrapInfo(resource, obj))
		},
		DeleteFunc: func(obj interface{}) {
			if funcs.DeleteFunc == nil {
				return
			}
			data, ok := obj.(*minimal.MinimalObject)
			if !ok {
				stale, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					logger.Errorf("informer %q reported unknown object to be deleted (%T)", resource.Name(), obj)
					return
				}
				if stale.Obj == nil {
					logger.Errorf("informer %q reported no stale object to be deleted", resource.Name())
					return
				}
				data, ok = stale.Obj.(*minimal.MinimalObject)
				if !ok {
					logger.Errorf("informer %q reported unknown stale object to be deleted (%T)", resource.Name(), stale.Obj)
					return
				}
			}
			funcs.DeleteFunc(wrapInfo(resource, data))
		},
		UpdateFunc: func(old, new interface{}) {
			if funcs.UpdateFunc == nil {
				return
			}
			funcs.UpdateFunc(wrapInfo(resource, old), wrapInfo(resource, new))
		},
	}
}

func wrapInfo(resource Interface, obj interface{}) ObjectInfo {
	m := obj.(*minimal.MinimalObject)
	m.SetGroupVersionKind(resource.GroupVersionKind())
	return &minimalObjectInfo{
		minimalObject: m,
		cluster:       resource.GetCluster(),
	}
}

type minimalObjectInfo struct {
	minimalObject *minimal.MinimalObject
	cluster       Cluster
}

func (this *minimalObjectInfo) Key() ObjectKey {
	m := this.minimalObject
	return NewKey(m.GroupVersionKind().GroupKind(), m.GetNamespace(), m.GetName())
}

func (this *minimalObjectInfo) Description() string {
	return this.Key().String()
}

func (this *minimalObjectInfo) GetResourceVersion() string {
	return this.minimalObject.ResourceVersion
}

func (this *minimalObjectInfo) GetCluster() Cluster {
	return this.cluster
}
