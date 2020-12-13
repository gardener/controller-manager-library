/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package module

import (
	"fmt"

	"github.com/gardener/controller-manager-library/pkg/certs"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/extension"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/module/handler"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/module/mappings"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"github.com/gardener/controller-manager-library/pkg/utils"
)

type module struct {
	extension.ElementBase
	extension.SharedAttributes

	definition  Definition
	env         Environment
	clusters    cluster.Clusters
	cluster     cluster.Interface
	handlers    map[string]handler.Interface
	certificate certs.CertificateSource

	config *ModuleConfig
}

func NewModule(env Environment, def Definition, cmp mappings.Definition) (*module, error) {
	options := env.GetConfig().GetSource(def.Name()).(*ModuleConfig)

	this := &module{
		definition: def,
		config:     options,
		env:        env,
		handlers:   map[string]handler.Interface{},
	}

	this.ElementBase = extension.NewElementBase(env.GetContext(), ctx_module, this, def.Name(), MODULE_SET_PREFIX, options)
	this.SharedAttributes = extension.NewSharedAttributes(this.ElementBase)

	required := cluster.Canonical(def.RequiredClusters())
	if len(required) != 0 {
		clusters, err := mappings.MapClusters(env.GetClusters(), cmp, required...)
		if err != nil {
			return nil, err
		}
		this.Infof("  using clusters %+v: %s (selected from %s)", required, clusters, env.GetClusters())
		this.clusters = clusters
		this.cluster = clusters.GetCluster(required[0])
	}
	return this, nil
}

func (this *module) GetEnvironment() Environment {
	return this.env
}

func (this *module) GetDefinition() Definition {
	return this.definition
}

func (this *module) GetClusterById(id string) cluster.Interface {
	return this.clusters.GetById(id)
}

func (this *module) GetCluster(name string) cluster.Interface {
	if name == CLUSTER_MAIN || name == "" {
		return this.GetMainCluster()
	}
	return this.clusters.GetCluster(name)
}

func (this *module) GetMainCluster() cluster.Interface {
	return this.cluster
}

func (this *module) GetClusterAliases(eff string) utils.StringSet {
	return this.clusters.GetAliases(eff)
}

func (this *module) GetEffectiveCluster(eff string) cluster.Interface {
	return this.clusters.GetEffective(eff)
}

func (this *module) GetObject(key resources.ClusterObjectKey) (resources.Object, error) {
	return this.clusters.GetObject(key)
}

func (this *module) GetCachedObject(key resources.ClusterObjectKey) (resources.Object, error) {
	return this.clusters.GetCachedObject(key)
}

func (this *module) handleSetup() error {
	this.Infof("setup module %s", this.definition.Name())
	for n, t := range this.definition.Handlers() {
		h, err := t(this)
		if err != nil {
			return err
		}
		this.handlers[n] = h
		if s, ok := h.(handler.SetupInterface); ok {
			this.Infof("  setup handler %s", n)
			err := s.Setup()
			if err != nil {
				return fmt.Errorf("setup of server %s handler %s failed: %s", this.definition.Name(), n, err)
			}
		}
	}
	return nil
}

func (this *module) handleStart() error {
	this.Infof("starting module %s", this.definition.Name())
	for n, h := range this.handlers {
		if s, ok := h.(handler.StartInterface); ok {
			this.Infof("  start handler %s", n)
			err := s.Start()
			if err != nil {
				return fmt.Errorf("start of module %s handler %s failed: %s", this.definition.Name(), n, err)
			}
		}
	}
	return nil
}
