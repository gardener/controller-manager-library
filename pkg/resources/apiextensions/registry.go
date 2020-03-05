/*
 * Copyright 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */

package apiextensions

import (
	"sync"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"github.com/gardener/controller-manager-library/pkg/utils"
)

type WebhookClientConfigProvider interface {
	GetClientConfig(gk schema.GroupKind, cluster resources.Cluster) WebhookClientConfigSource
}

type Registry interface {
	RegisterCRD(spec CRDSpecification) error
	GetCRD(gk schema.GroupKind) *CustomResourceDefinition

	RegisterClientConfigProvider(provider WebhookClientConfigProvider)
	GetClientConfig(gk schema.GroupKind, cluster resources.Cluster) WebhookClientConfigSource
}

type _registry struct {
	registry              map[schema.GroupKind]*CustomResourceDefinition
	lock                  sync.Mutex
	clientConfigProviders []WebhookClientConfigProvider
}

func NewRegistry() Registry {
	return &_registry{registry: map[schema.GroupKind]*CustomResourceDefinition{}}
}

func (this *_registry) RegisterCRD(spec CRDSpecification) error {
	crd, err := CRDObject(spec)
	if err != nil {
		return err
	}
	this.lock.Lock()
	defer this.lock.Unlock()
	logger.Infof("found crd specification %s: %+v", crd.CRDGroupKind(), crd.CRDVersions())
	this.registry[crd.CRDGroupKind()] = crd
	return nil
}

func (this *_registry) GetCRD(gk schema.GroupKind) *CustomResourceDefinition {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.registry[gk]
}

func (this *_registry) RegisterClientConfigProvider(provider WebhookClientConfigProvider) {
	if provider != nil {
		this.lock.Lock()
		defer this.lock.Unlock()
		this.clientConfigProviders = append(this.clientConfigProviders, provider)
	}
}

func (this *_registry) GetClientConfig(gk schema.GroupKind, cluster resources.Cluster) WebhookClientConfigSource {
	this.lock.Lock()
	defer this.lock.Unlock()
	for _, p := range this.clientConfigProviders {
		cfg := p.GetClientConfig(gk, cluster)
		if cfg != nil {
			return cfg
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

var registry = NewRegistry()

func RegisterCRD(spec CRDSpecification) error {
	return registry.RegisterCRD(spec)
}

func GetCRD(gk schema.GroupKind) *CustomResourceDefinition {
	return registry.GetCRD(gk)
}

func MustRegisterCRD(spec CRDSpecification) {
	utils.Must(registry.RegisterCRD(spec))
}

func RegisterClientConfigProvider(provider WebhookClientConfigProvider) {
	registry.RegisterClientConfigProvider(provider)
}

func GetClientConfig(gk schema.GroupKind, cluster resources.Cluster) WebhookClientConfigSource {
	return registry.GetClientConfig(gk, cluster)
}
