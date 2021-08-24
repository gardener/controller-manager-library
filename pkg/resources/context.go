/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/Masterminds/semver/v3"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	restclient "k8s.io/client-go/rest"

	"github.com/gardener/controller-manager-library/pkg/resources/abstract"
	"github.com/gardener/controller-manager-library/pkg/resources/errors"
	"github.com/gardener/controller-manager-library/pkg/utils"
)

type SchemeSource interface {
	Scheme(infos *ResourceInfos) *runtime.Scheme
	Equivalent(SchemeSource) bool
	String() string
}

type ResourceContext interface {
	abstract.ResourceContext

	GetResourceInfos() *ResourceInfos
	GetResourceInfosFor(gv schema.GroupVersion) []*Info
	Cluster

	GetParameterCodec() runtime.ParameterCodec
	GetClient(gv schema.GroupVersion) (restclient.Interface, error)

	SharedInformerFactory() SharedInformerFactory

	GetPreferred(gk schema.GroupKind) (*Info, error)
	Get(gvk schema.GroupVersionKind) (*Info, error)
}

type resourceContext struct {
	*ResourceInfos
	*Clients
	Cluster

	*abstract.AbstractResourceContext

	defaultResync         time.Duration
	sharedInformerFactory *sharedInformerFactory
}

func NewResourceContext(ctx context.Context, c Cluster, schemeSpec interface{}, defaultResync time.Duration) (ResourceContext, error) {
	res, err := NewResourceInfos(c)
	if err != nil {
		return nil, err
	}
	var scheme *runtime.Scheme
	if !utils.IsNil(schemeSpec) {
		switch s := schemeSpec.(type) {
		case *runtime.Scheme:
			scheme = s
		case SchemeSource:
			scheme = s.Scheme(res)
		default:
			return nil, fmt.Errorf("unknown scheme spec: %v", schemeSpec)
		}
	}
	rc := &resourceContext{
		Cluster:       c,
		ResourceInfos: res,
		defaultResync: defaultResync,
	}
	rc.AbstractResourceContext = abstract.NewAbstractResourceContext(ctx, rc, scheme, factory{})
	rc.Clients = NewClients(c.Config(), rc.Scheme())

	return rc, nil
}

func (c *resourceContext) GetServerVersion() *semver.Version {
	return c.ResourceInfos.GetServerVersion()
}

func (c *resourceContext) GetGroups() []schema.GroupVersion {
	return c.ResourceInfos.GetGroups()
}

func (c *resourceContext) Resources() Resources {
	c.SharedInformerFactory()
	return c.AbstractResourceContext.Resources().(Resources)
}

func (c *resourceContext) GetGVK(obj runtime.Object) (schema.GroupVersionKind, error) {
	var empty schema.GroupVersionKind

	gvks, _, err := c.ObjectKinds(obj)
	if err != nil {
		return empty, err
	}
	switch len(gvks) {
	case 0:
		return empty, errors.ErrUnknownResource.New("resource object type", reflect.TypeOf(obj))
	case 1:
		return gvks[0], nil
	default:
		for _, gvk := range gvks {
			def, err := c.GetPreferred(gvk.GroupKind())
			if err != nil {
				return empty, err
			}
			if def.Version() == gvk.Version {
				return gvk, nil
			}
		}
	}
	return empty, errors.New(errors.ERR_NON_UNIQUE_MAPPING, "non unique mapping for %T", obj)
}

// NewSharedInformerFactory constructs a new instance of sharedInformerFactory for all namespaces.
func (c *resourceContext) SharedInformerFactory() SharedInformerFactory {
	c.AbstractResourceContext.Lock()
	defer c.AbstractResourceContext.Unlock()

	if c.sharedInformerFactory == nil {
		c.sharedInformerFactory = newSharedInformerFactory(c, c.defaultResync)
	}
	return c.sharedInformerFactory
}
