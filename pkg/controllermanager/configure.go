/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package controllermanager

import (
	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/extension"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ConfigurationModifier func(c Configuration) Configuration

type Configuration struct {
	name          string
	description   string
	extension_reg extension.ExtensionRegistry
	cluster_reg   cluster.Registry
	configState

	globalMinimalWatch  map[schema.GroupKind]struct{}
	clusterMinimalWatch map[string]map[schema.GroupKind]struct{}
}

type configState struct {
	previous *configState
}

func (this *configState) pushState() {
	save := *this
	this.previous = &save
}

var _ cluster.RegistrationInterface = &Configuration{}

func Configure(name, desc string, scheme *runtime.Scheme) Configuration {
	return Configuration{
		name:          name,
		description:   desc,
		extension_reg: extension.NewExtensionRegistry(),
		cluster_reg:   cluster.NewRegistry(scheme),
		configState:   configState{},

		globalMinimalWatch:  map[schema.GroupKind]struct{}{},
		clusterMinimalWatch: map[string]map[schema.GroupKind]struct{}{},
	}
}

func (this Configuration) With(modifier ...ConfigurationModifier) Configuration {
	save := this.configState
	result := this
	for _, m := range modifier {
		result = m(result)
	}
	result.configState = save
	return result
}

func (this Configuration) Restore() Configuration {
	if &this.configState != nil {
		this.configState = *this.configState.previous
	}
	return this
}

func (this Configuration) ByDefault() Configuration {
	this.extension_reg = extension.DefaultRegistry()
	this.cluster_reg = cluster.DefaultRegistry()
	return this
}

func (this Configuration) GlobalMinimalWatch(groupKinds ...schema.GroupKind) Configuration {
	for _, gk := range groupKinds {
		this.globalMinimalWatch[gk] = struct{}{}
	}
	return this
}

func (this Configuration) MinimalWatch(clusterName string, groupKinds ...schema.GroupKind) Configuration {
	m, ok := this.clusterMinimalWatch[clusterName]
	if !ok {
		m = map[schema.GroupKind]struct{}{}
		this.clusterMinimalWatch[clusterName] = m
	}
	for _, gk := range groupKinds {
		m[gk] = struct{}{}
	}
	return this
}

func (this Configuration) RegisterExtension(reg extension.ExtensionType) {
	this.extension_reg.RegisterExtension(reg)
}
func (this Configuration) Extension(name string) extension.ExtensionType {
	for _, e := range this.extension_reg.GetExtensionTypes() {
		if e.Name() == name {
			return e
		}
	}
	return nil
}
func (this Configuration) RegisterCluster(reg cluster.Registerable) error {
	return this.cluster_reg.RegisterCluster(reg)
}
func (this Configuration) MustRegisterCluster(reg cluster.Registerable) cluster.RegistrationInterface {
	return this.cluster_reg.MustRegisterCluster(reg)
}

func (this Configuration) Definition() *Definition {
	cluster_defs := this.cluster_reg.GetDefinitions()
	for _, name := range cluster_defs.ClusterNames() {
		def := cluster_defs.Get(name).(cluster.InternalDefinition)
		for gk := range this.globalMinimalWatch {
			def.SetMinimalWatch(gk)
		}
		if m, ok := this.clusterMinimalWatch[name]; ok {
			for gk := range m {
				def.SetMinimalWatch(gk)
			}
		}
	}
	return &Definition{
		name:         this.name,
		description:  this.description,
		extensions:   this.extension_reg.GetDefinitions(),
		cluster_defs: cluster_defs,
	}
}
