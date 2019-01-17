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

package infrastructure

import (
	"context"
	"fmt"

	gardenextensionsv1alpha1 "github.com/gardener/controller-manager-library/pkg/apis/gardenextensions/v1alpha1"
	gardenextensionsinformersv1alpha1 "github.com/gardener/controller-manager-library/pkg/client/gardenextensions/informers/externalversions/gardenextensions/v1alpha1"
	gardenextensionslisters "github.com/gardener/controller-manager-library/pkg/client/gardenextensions/listers/gardenextensions/v1alpha1"
	typedclientset "github.com/gardener/controller-manager-library/pkg/clientsets/gardenextensions"
	"github.com/gardener/controller-manager-library/pkg/informerfactories"
	typedinformerfactories "github.com/gardener/controller-manager-library/pkg/informerfactories/gardenextensions"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"github.com/gardener/controller-manager-library/pkg/resources/gardenextensions"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

const kind = "Infrastructure"

var resourceGroupKind = gardenextensions.NewExtension(kind)

func init() {
	resources.MustRegister(&gardenextensionsv1alpha1.Infrastructure{}, &Type{resourceGroupKind})
}

type Handler struct {
	cluster  resources.Cluster
	informer gardenextensionsinformersv1alpha1.InfrastructureInformer
	lister   gardenextensionslisters.InfrastructureLister
}

var _ resources.Interface = &Handler{}

func ObjectAsResource(obj *gardenextensionsv1alpha1.Infrastructure, cluster resources.Cluster) resources.Object {
	r := &Resource{nil, resourceGroupKind, obj}
	r.ResourceBase = resources.NewResourceBase(r, cluster)
	return r
}

func (h *Handler) GetResource(obj interface{}) (resources.Object, error) {
	switch o := obj.(type) {
	case *gardenextensionsv1alpha1.Infrastructure:
		return ObjectAsResource(o, h.cluster), nil
	case resources.Key:
		if o.GroupKind() != resourceGroupKind.GetGroupKind() {
			return nil, fmt.Errorf("%s cannot handle group/kind '%s'", kind, o.GroupKind())
		}
		obj, err := h.lister.Infrastructures(o.Namespace()).Get(o.Name())
		if err != nil {
			return nil, err
		}
		return ObjectAsResource(obj, h.cluster), nil
	default:
		return nil, fmt.Errorf("unsupported type '%T' for source object", obj)
	}
}

func (h *Handler) AddEventHandler(handlers cache.ResourceEventHandlerFuncs) error {
	logger.Infof("adding handler for %s", kind)
	h.informer.Informer().AddEventHandler(handlers)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////

type Type struct {
	resources.GroupKind
}

var _ resources.Type = &Type{}

func (t *Type) NewHandler(ctx context.Context, cluster resources.Cluster) (resources.Interface, error) {
	factory, err := typedinformerfactories.SharedInformerFactory(cluster.InformerFactories())
	if err != nil {
		return nil, err
	}

	var (
		informer = factory.Gardenextensions().V1alpha1().Infrastructures()
		handler  = &Handler{
			cluster:  cluster,
			lister:   informer.Lister(),
			informer: informer,
		}
	)

	if err := informerfactories.Start(ctx, factory, informer.Informer().HasSynced); err != nil {
		return nil, err
	}

	return handler, nil
}

type Resource struct {
	*resources.ResourceBase
	resources.GroupKind
	*gardenextensionsv1alpha1.Infrastructure
}

var _ gardenextensions.ExtensionInterface = &Resource{}
var _ resources.Object = &Resource{}

func (i *Resource) GetObject() runtime.Object {
	return i.Infrastructure
}

func (i *Resource) GetExtensionType() string {
	return i.Spec.Type
}

func (i *Resource) DeepCopy() resources.Object {
	r := &Resource{nil, resourceGroupKind, i.Infrastructure.DeepCopy()}
	r.ResourceBase = i.ResourceBase.For(r)
	return r
}

func (i *Resource) UpdateIn(cluster resources.Cluster) error {
	clientset, err := typedclientset.Clientset(cluster.Clientsets())
	if err != nil {
		return err
	}
	_, err = clientset.GardenextensionsV1alpha1().Infrastructures(i.GetNamespace()).Update(i.Infrastructure)
	return err
}

func (i *Resource) Update() error {
	clientset, err := typedclientset.Clientset(i.GetCluster().Clientsets())
	if err != nil {
		return err
	}
	_, err = clientset.GardenextensionsV1alpha1().Infrastructures(i.GetNamespace()).Update(i.Infrastructure)
	return err
}
