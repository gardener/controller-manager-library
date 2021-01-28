/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"fmt"
	"reflect"

	"github.com/gardener/controller-manager-library/pkg/logger"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type _resource struct {
	AbstractResource
	info   *Info
	client restclient.Interface

	handlers map[interface{}]cache.ResourceEventHandler
}

var _ Interface = &_resource{}

type namespacedResource struct {
	resource  *AbstractResource
	namespace string
	lister    NamespacedLister
}

/////////////////////////////////////////////////////////////////////////////////

func newResource(ctx ResourceContext, otype, ltype reflect.Type, gvk schema.GroupVersionKind) (*_resource, error) {
	info, err := ctx.Get(gvk)
	if err != nil {
		return nil, err
	}

	client, err := ctx.GetClient(gvk.GroupVersion())
	if err != nil {
		return nil, err
	}

	if otype == nil {
		otype = unstructuredType
	}
	r := &_resource{
		info:     info,
		client:   client,
		handlers: map[interface{}]cache.ResourceEventHandler{},
	}
	r.AbstractResource, _ = NewAbstractResource(ctx, new_i_resource(r), otype, ltype, gvk)
	return r, nil
}

func (this *_resource) GetCluster() Cluster {
	return this.ResourceContext().GetCluster()
}

func (this *_resource) ResourceContext() ResourceContext {
	return this.AbstractResource.ResourceContext().(ResourceContext)
}

func (this *_resource) Resources() Resources {
	return this.ResourceContext().Resources()
}

var unstructuredType = reflect.TypeOf(unstructured.Unstructured{})

func (this *_resource) IsUnstructured() bool {
	return this.ObjectType() == unstructuredType
}

func (this *_resource) Info() *Info {
	return this.info
}

func (this *_resource) Client() restclient.Interface {
	return this.client
}

func (this *_resource) GetParameterCodec() runtime.ParameterCodec {
	return this.ResourceContext().GetParameterCodec()
}

func (this *_resource) AddRawEventHandler(handlers cache.ResourceEventHandler) error {
	return this.AddRawSelectedEventHandler(handlers, "", nil)
}

func (this *_resource) AddRawInfoEventHandler(handlers cache.ResourceEventHandler) error {
	return this.AddRawSelectedInfoEventHandler(handlers, "", nil)
}

func (this *_resource) AddRawSelectedEventHandler(handlers cache.ResourceEventHandler, namespace string, optionsFunc TweakListOptionsFunc) error {
	return this.addRawSelectedEventHandler(false, handlers, namespace, optionsFunc)
}

func (this *_resource) AddRawSelectedInfoEventHandler(handlers cache.ResourceEventHandler, namespace string, optionsFunc TweakListOptionsFunc) error {
	return this.addRawSelectedEventHandler(true, handlers, namespace, optionsFunc)
}

func (this *_resource) addRawSelectedEventHandler(minimal bool, handlers cache.ResourceEventHandler, namespace string, optionsFunc TweakListOptionsFunc) error {
	withNamespace := "global"
	if namespace != "" {
		withNamespace = fmt.Sprintf("namespace %s", namespace)
	}
	logger.Infof("adding watch for %s (cluster %s, %s)", this.GroupVersionKind(), this.GetCluster().GetId(), withNamespace)
	informer, err := this.helper.Internal.I_getInformer(minimal, namespace, optionsFunc)
	if err != nil {
		return err
	}
	informer.AddEventHandler(handlers)
	return nil
}

func (this *_resource) RemoveRawEventHandler(handlers cache.ResourceEventHandler) error {
	return this.RemoveRawSelectedEventHandler(handlers, "", nil)
}

func (this *_resource) RemoveRawInfoEventHandler(handlers cache.ResourceEventHandler) error {
	return this.RemoveRawSelectedInfoEventHandler(handlers, "", nil)
}

func (this *_resource) RemoveRawSelectedEventHandler(handlers cache.ResourceEventHandler, namespace string, optionsFunc TweakListOptionsFunc) error {
	return this.removeRawSelectedEventHandler(false, handlers, namespace, optionsFunc)
}

func (this *_resource) RemoveRawSelectedInfoEventHandler(handlers cache.ResourceEventHandler, namespace string, optionsFunc TweakListOptionsFunc) error {
	return this.removeRawSelectedEventHandler(true, handlers, namespace, optionsFunc)
}

func (this *_resource) removeRawSelectedEventHandler(minimal bool, handler cache.ResourceEventHandler, namespace string, optionsFunc TweakListOptionsFunc) error {
	withNamespace := "global"
	if namespace != "" {
		withNamespace = fmt.Sprintf("namespace %s", namespace)
	}
	logger.Infof("adding watch for %s (cluster %s, %s)", this.GroupVersionKind(), this.GetCluster().GetId(), withNamespace)
	informer, err := this.helper.Internal.I_getInformer(minimal, namespace, optionsFunc)
	if err != nil {
		return err
	}
	return informer.RemoveEventHandler(handler)
}

func (this *_resource) AddEventHandler(handler ResourceEventHandler) error {
	h, _ := this.mapHandler(handler)
	return this.AddRawEventHandler(h)
}

func (this *_resource) AddSelectedEventHandler(handler ResourceEventHandler, namespace string, optionsFunc TweakListOptionsFunc) error {
	h, _ := this.mapHandler(handler)
	return this.AddRawSelectedEventHandler(h, namespace, optionsFunc)
}

func (this *_resource) AddInfoEventHandler(handler ResourceInfoEventHandler) error {
	h, _ := this.mapInfoHandler(handler)
	return this.AddRawInfoEventHandler(h)
}

func (this *_resource) AddSelectedInfoEventHandler(handler ResourceInfoEventHandler, namespace string, optionsFunc TweakListOptionsFunc) error {
	h, _ := this.mapInfoHandler(handler)
	return this.AddRawSelectedInfoEventHandler(h, namespace, optionsFunc)
}

func (this *_resource) RemoveEventHandler(handler ResourceEventHandler) error {
	h, removable := this.mapHandler(handler)
	if !removable {
		return fmt.Errorf("handler is not removable")
	}
	return this.RemoveRawEventHandler(h)
}

func (this *_resource) RemoveSelectedEventHandler(handler ResourceEventHandler, namespace string, optionsFunc TweakListOptionsFunc) error {
	h, removable := this.mapHandler(handler)
	if !removable {
		return fmt.Errorf("handler is not removable")
	}
	return this.AddRawSelectedEventHandler(h, namespace, optionsFunc)
}

func (this *_resource) RemoveInfoEventHandler(handler ResourceInfoEventHandler) error {
	h, removable := this.mapInfoHandler(handler)
	if !removable {
		return fmt.Errorf("handler is not removable")
	}
	return this.AddRawInfoEventHandler(h)
}

func (this *_resource) RemoveSelectedInfoEventHandler(handler ResourceInfoEventHandler, namespace string, optionsFunc TweakListOptionsFunc) error {
	h, removable := this.mapInfoHandler(handler)
	if !removable {
		return fmt.Errorf("handler is not removable")
	}
	return this.AddRawSelectedInfoEventHandler(h, namespace, optionsFunc)
}

func (this *_resource) NormalEventf(name ObjectDataName, reason, msgfmt string, args ...interface{}) {
	this.Resources().Eventf(this.CreateData(name), v1.EventTypeNormal, reason, msgfmt, args...)
}

func (this *_resource) WarningEventf(name ObjectDataName, reason, msgfmt string, args ...interface{}) {
	this.Resources().Eventf(this.CreateData(name), v1.EventTypeWarning, reason, msgfmt, args...)
}

func (this *_resource) namespacedRequest(req *restclient.Request, namespace string) *restclient.Request {
	return req.NamespaceIfScoped(namespace, this.Namespaced()).Resource(this.Name())
}

func (this *_resource) resourceRequest(req *restclient.Request, obj ObjectDataName, sub ...string) *restclient.Request {
	if this.Namespaced() && obj != nil {
		req = req.Namespace(obj.GetNamespace())
	}
	return req.Resource(this.Name()).SubResource(sub...)
}

func (this *_resource) objectRequest(req *restclient.Request, obj ObjectDataName, sub ...string) *restclient.Request {
	return this.resourceRequest(req, obj, sub...).Name(obj.GetName())
}
