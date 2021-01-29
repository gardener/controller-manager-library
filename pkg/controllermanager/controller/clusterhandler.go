/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"github.com/gardener/controller-manager-library/pkg/utils"
)

type poolInfo struct {
	useCount int
}

type clusterResourceInfo struct {
	resource    resources.Interface
	pools       map[*pool]*poolInfo
	minimal     bool
	namespace   string
	optionsFunc resources.TweakListOptionsFunc
}

func (this *clusterResourceInfo) Check(minimal bool, namespace string, optionsFunc resources.TweakListOptionsFunc) error {
	if this.minimal != minimal {
		return fmt.Errorf("watch namespace minimal mismatch for resource %s (%q != %q)", GetResourceKey(this.resource), this.minimal, minimal)
	}
	if this.namespace != namespace {
		return fmt.Errorf("watch namespace mismatch for resource %s (%q != %q)", GetResourceKey(this.resource), this.namespace, namespace)
	}
	if (this.optionsFunc == nil) != (optionsFunc == nil) {
		return fmt.Errorf("watch options mismatch for resource %s", GetResourceKey(this.resource))
	}
	return nil
}

func (this *clusterResourceInfo) IsUsed() bool {
	for _, i := range this.pools {
		if i.useCount > 0 {
			return true
		}
	}
	return false
}

func (this *clusterResourceInfo) UsePool(usedpool *pool) {
	if i := this.pools[usedpool]; i != nil {
		i.useCount++
	} else {
		this.pools[usedpool] = &poolInfo{1}
	}
}

func (this *clusterResourceInfo) List() ([]resources.Object, error) {
	opts := v1.ListOptions{}
	if this.optionsFunc != nil {
		this.optionsFunc(&opts)
	}
	if this.namespace != "" {
		return this.resource.Namespace(this.namespace).List(opts)
	}
	return this.resource.List(opts)
}

type WatchKey struct {
	rtype        ResourceKey
	unstructured bool
	minimal      bool
}

func (this WatchKey) ResourceKey() ResourceKey { return this.rtype }
func (this WatchKey) Unstructured() bool       { return this.unstructured }
func (this WatchKey) Minimal() bool            { return this.minimal }

type ClusterHandler struct {
	logger.LogContext
	controller *controller
	cluster    cluster.Interface
	resources  map[ResourceKey]*clusterResourceInfo
	cache      sync.Map

	regularHandler resources.ResourceEventHandler
	infoHandler    resources.ResourceInfoEventHandler
}

func newClusterHandler(controller *controller, cluster cluster.Interface) (*ClusterHandler, error) {
	c := &ClusterHandler{
		LogContext: controller.NewContext("cluster", cluster.GetName()),
		controller: controller,
		cluster:    cluster,
		resources:  map[ResourceKey]*clusterResourceInfo{},
	}

	c.regularHandler = &resources.ResourceEventHandlerFuncs{
		AddFunc:    c.objectAdd,
		UpdateFunc: c.objectUpdate,
		DeleteFunc: c.objectDelete,
	}

	c.infoHandler = &resources.ResourceInfoEventHandlerFuncs{
		AddFunc:    c.objectInfoAdd,
		UpdateFunc: c.objectInfoUpdate,
		DeleteFunc: c.objectInfoDelete,
	}
	return c, nil
}

func (c *ClusterHandler) whenReady() {
	c.controller.whenReady()
}

func (c *ClusterHandler) String() string {
	return c.cluster.GetName()
}

func (c *ClusterHandler) GetAliases() utils.StringSet {
	return c.controller.GetClusterAliases(c.cluster.GetName())
}

func (c *ClusterHandler) GetResource(resourceKey ResourceKey) (resources.Interface, error) {
	return c.cluster.GetResource(resourceKey.GroupKind())
}

func (c *ClusterHandler) newResourceInfo(key WatchKey, resource resources.Interface, namespace string, optionsFunc resources.TweakListOptionsFunc) *clusterResourceInfo {
	i := &clusterResourceInfo{
		pools:       map[*pool]*poolInfo{},
		namespace:   namespace,
		optionsFunc: optionsFunc,
		resource:    resource,
		minimal:     key.Minimal(),
	}
	c.resources[key.ResourceKey()] = i
	return i
}

func (c *ClusterHandler) unregister(watchResource WatchResource, namespace string, optionsFunc resources.TweakListOptionsFunc, usedpool *pool) error {
	watchKey := watchResource.WatchKey()
	resourceKey := watchResource.ResourceType()
	watchKey.minimal = watchResource.ShouldEnforceMinimal() || c.cluster.Definition().IsMinimalWatchEnforced(resourceKey.GroupKind())
	i := c.resources[resourceKey]
	if i != nil && i.pools[usedpool] != nil && i.pools[usedpool].useCount > 0 {
		if err := i.Check(watchKey.Minimal(), namespace, optionsFunc); err != nil {
			return err
		}
		i.pools[usedpool].useCount--
		if i.pools[usedpool].useCount == 0 {
			if watchKey.Minimal() {
				i.resource.RemoveSelectedInfoEventHandler(c.infoHandler, namespace, optionsFunc)
			} else {
				i.resource.RemoveSelectedEventHandler(c.regularHandler, namespace, optionsFunc)
			}
			delete(c.resources, resourceKey)
		}
	}
	return nil
}

func (c *ClusterHandler) register(watchResource WatchResource, namespace string, optionsFunc resources.TweakListOptionsFunc, usedpool *pool) error {
	watchKey := watchResource.WatchKey()
	resourceKey := watchResource.ResourceType()
	watchKey.minimal = watchResource.ShouldEnforceMinimal() || c.cluster.Definition().IsMinimalWatchEnforced(resourceKey.GroupKind())
	i := c.resources[resourceKey]
	if i == nil {
		resource, err := c.cluster.GetResource(resourceKey.GroupKind(), watchResource.Unstructured())
		if err != nil {
			return err
		}
		i = c.newResourceInfo(watchKey, resource, namespace, optionsFunc)
	}

	if err := i.Check(watchKey.Minimal(), namespace, optionsFunc); err != nil {
		return err
	}
	if !i.IsUsed() {
		if watchKey.Minimal() {
			if err := i.resource.AddSelectedInfoEventHandler(c.infoHandler, namespace, optionsFunc); err != nil {
				return err
			}
		} else {
			if err := i.resource.AddSelectedEventHandler(c.regularHandler, namespace, optionsFunc); err != nil {
				return err
			}
		}
	} else {
		if optionsFunc != nil {
			opts1 := &v1.ListOptions{}
			opts2 := &v1.ListOptions{}
			i.optionsFunc(opts1)
			optionsFunc(opts2)
			if !reflect.DeepEqual(opts1, opts2) {
				return fmt.Errorf("watch options mismatch for resource %s (%+v != %+v)", resourceKey, opts1, opts2)
			}
		}
	}
	i.UsePool(usedpool)

	return nil
}

///////////////////////////////////////////////////////////////////////////////

func (c *ClusterHandler) EnqueueKey(key resources.ClusterObjectKey) error {
	// c.Infof("enqueue %s", obj.Description())
	gk := key.GroupKind()
	rk := NewResourceKey(gk.Group, gk.Kind)
	i := c.resources[rk]
	if i == nil {
		return fmt.Errorf("cluster %q: no resource info for %s", c, rk)
	}
	if i.pools == nil || len(i.pools) == 0 {
		return fmt.Errorf("cluster %q: no worker pool for type %s", c, rk)
	}
	for p := range i.pools {
		p.EnqueueKey(key)
	}
	return nil
}

func (c *ClusterHandler) enqueue(obj resources.ObjectInfo, e func(p *pool, r resources.ObjectInfo)) error {
	c.whenReady()
	// c.Infof("enqueue %s", obj.Description())
	i := c.resources[GetResourceKey(obj)]
	if i == nil {
		c.Infof("@@@ trying to enqueue %s for cluster %a", obj.Key(), c.String())
		return nil
	}
	if i.pools == nil || len(i.pools) == 0 {
		return fmt.Errorf("no worker pool for type %s", obj.Key().GroupKind())
	}
	for p := range i.pools {
		// p.Infof("enqueue %s", resources.ObjectrKey(obj))
		e(p, obj)
	}
	return nil
}

func enq(p *pool, obj resources.ObjectInfo) {
	p.EnqueueObject(obj)
}

func (c *ClusterHandler) EnqueueObject(obj resources.ObjectInfo) error {
	return c.enqueue(obj, enq)
}

func enqRateLimited(p *pool, obj resources.ObjectInfo) {
	p.EnqueueObjectRateLimited(obj)
}
func (c *ClusterHandler) EnqueueObjectRateLimited(obj resources.ObjectInfo) error {
	return c.enqueue(obj, enqRateLimited)
}

func (c *ClusterHandler) EnqueueObjectAfter(obj resources.ObjectInfo, duration time.Duration) error {
	e := func(p *pool, obj resources.ObjectInfo) {
		p.EnqueueObjectAfter(obj, duration)
	}
	return c.enqueue(obj, e)
}

///////////////////////////////////////////////////////////////////////////////

func (c *ClusterHandler) GetObject(key resources.ClusterObjectKey) (resources.Object, error) {
	o, ok := c.cache.Load(key.ObjectKey())
	if o == nil || !ok {
		return nil, nil
	}
	if obj, ok := o.(resources.Object); ok {
		return obj, nil
	}

	resource, err := c.cluster.GetResource(key.GroupKind())
	if err != nil {
		return nil, err
	}
	obj, err := resource.Get(key.ObjectKey())
	if err != nil && errors.IsNotFound(err) {
		return nil, nil
	}
	return obj, err
}

func (c *ClusterHandler) objectAdd(obj resources.Object) {
	if c.controller.mustHandle(obj) {
		c.objectInfoAdd(obj)
	}
}

func (c *ClusterHandler) objectUpdate(old, new resources.Object) {
	if !c.controller.mustHandle(old) && !c.controller.mustHandle(new) {
		return
	}
	c.objectInfoUpdate(old, new)
}

func (c *ClusterHandler) objectDelete(obj resources.Object) {
	if c.controller.mustHandle(obj) {
		c.objectInfoDelete(obj)
	}
}

func (c *ClusterHandler) objectInfoAdd(obj resources.ObjectInfo) {
	c.Debugf("** GOT add event for %s", obj.Description())

	c.cache.Store(obj.Key(), obj)
	c.EnqueueObject(obj)
}

func (c *ClusterHandler) objectInfoUpdate(old, new resources.ObjectInfo) {
	c.Debugf("** GOT update event for %s: %s", new.Description(), new.GetResourceVersion())
	c.cache.Store(new.Key(), new)
	c.EnqueueObject(new)
}

func (c *ClusterHandler) objectInfoDelete(obj resources.ObjectInfo) {
	c.Debugf("** GOT delete event for %s: %s", obj.Description(), obj.GetResourceVersion())

	c.cache.Delete(obj.Key())
	c.EnqueueObject(obj)
}
