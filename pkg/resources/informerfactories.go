/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/gardener/controller-manager-library/pkg/resources/errors"

	"github.com/gardener/controller-manager-library/pkg/kutil"
	"github.com/gardener/controller-manager-library/pkg/logger"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type internalInformerFactory interface {
	informerFor(informerType reflect.Type, gvk schema.GroupVersionKind) (GenericInformer, error)

	Start(stopCh <-chan struct{})
	WaitForCacheSync(stopCh <-chan struct{})
}

///////////////////////////////////////////////////////////////////////////////
// Generic Informer Factory

type genericInformerFactory struct {
	lock sync.Mutex

	context     *resourceContext
	optionsFunc TweakListOptionsFunc
	namespace   string

	defaultResync time.Duration
	informers     map[schema.GroupVersionKind]GenericInformer
	// startedInformers is used for tracking which informers have been started.
	// This allows Start() to be called multiple times safely.
	startedInformers map[schema.GroupVersionKind]bool
}

var _ internalInformerFactory = &genericInformerFactory{}

func newGenericInformerFactory(rctx *resourceContext, defaultResync time.Duration, namespace string, optionsFunc TweakListOptionsFunc) *genericInformerFactory {
	return &genericInformerFactory{
		context:       rctx,
		defaultResync: defaultResync,
		optionsFunc:   optionsFunc,
		namespace:     namespace,

		informers:        make(map[schema.GroupVersionKind]GenericInformer),
		startedInformers: make(map[schema.GroupVersionKind]bool),
	}
}

// Start initializes all requested informers.
func (f *genericInformerFactory) Start(stopCh <-chan struct{}) {
	f.lock.Lock()
	defer f.lock.Unlock()

	for informerType, informer := range f.informers {
		if !f.startedInformers[informerType] {
			go informer.Run(stopCh)
			f.startedInformers[informerType] = true
		}
	}
}

// WaitForCacheSync waits for all started informers' cache were synced.
func (f *genericInformerFactory) WaitForCacheSync(stopCh <-chan struct{}) {
	informers := func() map[schema.GroupVersionKind]cache.SharedIndexInformer {
		f.lock.Lock()
		defer f.lock.Unlock()

		informers := map[schema.GroupVersionKind]cache.SharedIndexInformer{}
		for informerType, informer := range f.informers {
			if f.startedInformers[informerType] {
				informers[informerType] = informer
			}
		}
		return informers
	}()

	for _, informer := range informers {
		cache.WaitForCacheSync(stopCh, informer.HasSynced)
	}
}

func (f *genericInformerFactory) informerFor(informerType reflect.Type, gvk schema.GroupVersionKind) (GenericInformer, error) {
	f.lock.Lock()
	defer f.lock.Unlock()

	informer, exists := f.informers[gvk]
	if exists {
		return informer, nil
	}

	l := kutil.DetermineListType(f.context.Scheme(), gvk.GroupVersion(), informerType)
	if l == nil {
		return nil, errors.New(errors.ERR_NO_LIST_TYPE, "no list type found for %s", informerType)
	}

	client, err := f.getClient(gvk.GroupVersion())
	if err != nil {
		return nil, err
	}

	info, err := f.context.Get(gvk)
	if err != nil {
		return nil, err
	}
	informer = f.newInformer(client, info, informerType, l)
	f.informers[gvk] = informer

	return informer, nil
}

func (f *genericInformerFactory) queryInformerFor(informerType reflect.Type, gvk schema.GroupVersionKind) GenericInformer {
	f.lock.Lock()
	defer f.lock.Unlock()

	informer, exists := f.informers[gvk]
	if exists {
		return informer
	}
	return nil
}

func (f *genericInformerFactory) getClient(gv schema.GroupVersion) (restclient.Interface, error) {
	return f.context.GetClient(gv)
}

func (f *genericInformerFactory) newInformer(client restclient.Interface, res *Info, elemType reflect.Type, listType reflect.Type) GenericInformer {
	logger.Infof("new generic informer for %s (%s) %s (%d seconds)", elemType, res.GroupVersionKind(), listType, f.defaultResync/time.Second)
	indexers := cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}
	ctx := context.TODO()
	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				if f.optionsFunc != nil {
					f.optionsFunc(&options)
				}
				result := reflect.New(listType).Interface().(runtime.Object)
				r := client.Get().
					Resource(res.Name()).
					VersionedParams(&options, f.context.Clients.parametercodec)
				if res.Namespaced() {
					r = r.Namespace(f.namespace)
				}

				return result, r.Do(ctx).Into(result)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				options.Watch = true
				if f.optionsFunc != nil {
					f.optionsFunc(&options)
				}
				r := client.Get().
					Resource(res.Name()).
					VersionedParams(&options, f.context.Clients.parametercodec)
				if res.Namespaced() {
					r = r.Namespace(f.namespace)
				}

				w, err := r.Watch(ctx)
				return w, err
			},
		},
		reflect.New(elemType).Interface().(runtime.Object),
		f.defaultResync,
		indexers,
	)
	return &genericInformer{informer, res}
}
