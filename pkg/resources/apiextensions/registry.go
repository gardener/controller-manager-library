/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package apiextensions

import (
	"sync"

	"github.com/Masterminds/semver/v3"
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
	OverwriteCRD(k8svers *semver.Version, spec CRDSpecification) error
	GetCRDs(gk schema.GroupKind) *CustomResourceDefinitionVersions
	GetCRDDataFor(gk schema.GroupKind, cluster resources.Cluster) resources.ObjectData
	GetCRDObjectFor(gk schema.GroupKind, cluster resources.Cluster) resources.Object

	RegisterClientConfigProvider(provider WebhookClientConfigProvider)
	WebhookClientConfigProvider

	AddToRegistry(r Registry)
}

type _registry struct {
	registry              map[schema.GroupKind]*CustomResourceDefinitionVersions
	lock                  sync.Mutex
	clientConfigProviders []WebhookClientConfigProvider
}

func NewRegistry() Registry {
	return &_registry{registry: map[schema.GroupKind]*CustomResourceDefinitionVersions{}}
}

func (this *_registry) RegisterCRD(spec CRDSpecification) error {
	crds, err := NewDefaultedCustomResourceDefinitionVersions(spec)
	if err != nil {
		return err
	}
	this.lock.Lock()
	defer this.lock.Unlock()
	logger.InitInfof("found crd specification %s: %s", crds.GroupKind(), crds.GetDefault().CRDVersions())
	this.registry[crds.GroupKind()] = crds
	return nil
}

func (this *_registry) OverwriteCRD(version *semver.Version, spec CRDSpecification) error {
	crd, err := GetCustomResourceDefinition(spec)
	if err != nil {
		return err
	}
	this.lock.Lock()
	defer this.lock.Unlock()
	logger.Infof("found crd specification %s for kubernetes version %s: %+v", crd.CRDGroupKind(), version, crd.CRDVersions())
	crds := this.registry[crd.CRDGroupKind()]
	if crds == nil {
		crds = NewCustomResourceDefinitionVersions(crd.GroupVersionKind().GroupKind())
		this.registry[crd.CRDGroupKind()] = crds
	}
	crds.Override(version, crd)
	return nil
}

func (this *_registry) GetCRDs(gk schema.GroupKind) *CustomResourceDefinitionVersions {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.registry[gk]
}

func (this *_registry) GetCRDObjectFor(gk schema.GroupKind, cluster resources.Cluster) resources.Object {
	obj, err := cluster.Resources().Wrap(this.GetCRDDataFor(gk, cluster))
	utils.Must(err)
	return obj
}

func (this *_registry) GetCRDDataFor(gk schema.GroupKind, cluster resources.Cluster) resources.ObjectData {
	crds := this.GetCRDs(gk)
	if crds == nil {
		return nil
	}

	crd := crds.GetFor(cluster.GetServerVersion())
	if crd == nil {
		return nil
	}
	return crd.DataFor(cluster, this)
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

func (this *_registry) AddToRegistry(r Registry) {
	this.lock.Lock()
	defer this.lock.Unlock()
	for _, v := range this.registry {
		def := v.GetDefault()
		if def != nil {
			utils.Must(r.RegisterCRD(def))
		}
		for v, o := range v.GetVersions() {
			if v != nil {
				utils.Must(r.OverwriteCRD(v, o))
			}
		}
	}
	for _, p := range this.clientConfigProviders {
		r.RegisterClientConfigProvider(p)
	}
}

////////////////////////////////////////////////////////////////////////////////

var registry = NewRegistry()

func DefaultRegistry() Registry {
	return registry
}

func RegisterCRD(spec CRDSpecification) error {
	return registry.RegisterCRD(spec)
}

func GetCRDs(gk schema.GroupKind) *CustomResourceDefinitionVersions {
	return registry.GetCRDs(gk)
}

func GetCRDFor(gk schema.GroupKind, cluster resources.Cluster) resources.ObjectData {
	return registry.GetCRDDataFor(gk, cluster)
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
