/*
 * Copyright 2019 SAP SE or an SAP affiliate company. All rights reserved.
 * This file is licensed under the Apache Software License, v. 2 except as noted
 * otherwise in the LICENSE file
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

package webhook

import (
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"time"

	"github.com/gardener/controller-manager-library/pkg/config"
	"github.com/gardener/controller-manager-library/pkg/controllermanager"

	adminreg "k8s.io/api/admissionregistration/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type _Definition struct {
	name               string
	keys               []controllermanager.ResourceKey
	cluster            string
	scheme             *runtime.Scheme
	kind               string
	handler            AdmissionHandlerType
	namespaces         *meta.LabelSelector
	operations         []adminreg.OperationType
	policy             adminreg.FailurePolicyType
	configs            controllermanager.OptionDefinitions
	activateExplicitly bool
}

var _ Definition = &_Definition{}

func (this *_Definition) GetResources() []controllermanager.ResourceKey {
	return append(this.keys[:0:0], this.keys...)
}
func (this *_Definition) GetName() string {
	return this.name
}
func (this *_Definition) GetCluster() string {
	return this.cluster
}
func (this *_Definition) GetScheme() *runtime.Scheme {
	return this.scheme
}
func (this *_Definition) GetKind() string {
	return this.kind
}
func (this *_Definition) GetHandlerType() AdmissionHandlerType {
	return this.handler
}
func (this *_Definition) GetNamespaces() *meta.LabelSelector {
	return this.namespaces
}
func (this *_Definition) GetFailurePolicy() adminreg.FailurePolicyType {
	if this.policy == "" {
		return adminreg.Fail
	}
	return this.policy
}
func (this *_Definition) GetOperations() []adminreg.OperationType {
	result := this.operations[:0:0]
	copy(result, this.operations)
	return result
}
func (this *_Definition) ConfigOptions() map[string]OptionDefinition {
	cfgs := map[string]OptionDefinition{}
	for n, d := range this.configs {
		cfgs[n] = d
	}
	return cfgs
}
func (this *_Definition) ActivateExplicitly() bool {
	return this.activateExplicitly
}

func (this *_Definition) String() string {
	s := fmt.Sprintf("%s webhook %q:\n", this.kind, this.name)
	s += fmt.Sprintf("  cluster: %s\n", this.cluster)
	s += fmt.Sprintf("  gvks:\n")
	for _, k := range this.keys {
		s += fmt.Sprintf("  - %s\n", k)
	}
	s += fmt.Sprintf("  namespaces: %+v\n", this.namespaces)
	if this.scheme != nil {
		s += "  scheme set\n"
	}
	return s
}

////////////////////////////////////////////////////////////////////////////////
// Confihuration
////////////////////////////////////////////////////////////////////////////////

type Configuration struct {
	settings _Definition
}

func Configure(name string) Configuration {
	return Configuration{
		settings: _Definition{
			name:    name,
			configs: controllermanager.OptionDefinitions{},
		},
	}
}

func (this Configuration) Name(name string) Configuration {
	this.settings.name = name
	return this
}

func (this Configuration) Cluster(name string) Configuration {
	this.settings.cluster = name
	return this
}

func (this Configuration) Scheme(scheme *runtime.Scheme) Configuration {
	this.settings.scheme = scheme
	return this
}

func (this Configuration) Resource(group, kind string) Configuration {
	this.settings.keys = append(this.settings.keys, controllermanager.NewResourceKey(group, kind))
	return this
}

func (this Configuration) IgnoreFailures() Configuration {
	this.settings.policy = adminreg.Ignore
	return this
}

func (this Configuration) Operation(op ...adminreg.OperationType) Configuration {
	this.settings.operations = append(this.settings.operations, op...)
	return this
}

func (this Configuration) Namespaces(selector *meta.LabelSelector) Configuration {
	this.settings.namespaces = selector
	return this
}

func (this Configuration) Handler(htype AdmissionHandlerType) Configuration {
	this.settings.handler = htype
	return this
}

func (this Configuration) StringOption(name string, desc string) Configuration {
	return this.addOption(name, config.StringOption, "", desc)
}
func (this Configuration) DefaultedStringOption(name, def string, desc string) Configuration {
	return this.addOption(name, config.StringOption, def, desc)
}

func (this Configuration) StringArrayOption(name string, desc string) Configuration {
	return this.addOption(name, config.StringArrayOption, nil, desc)
}
func (this Configuration) DefaultedStringArrayOption(name string, def []string, desc string) Configuration {
	return this.addOption(name, config.StringArrayOption, def, desc)
}

func (this Configuration) IntOption(name string, desc string) Configuration {
	return this.addOption(name, config.IntOption, 0, desc)
}
func (this Configuration) DefaultedIntOption(name string, def int, desc string) Configuration {
	return this.addOption(name, config.IntOption, def, desc)
}

func (this Configuration) BoolOption(name string, desc string) Configuration {
	return this.addOption(name, config.BoolOption, false, desc)
}
func (this Configuration) DefaultedBoolOption(name string, def bool, desc string) Configuration {
	return this.addOption(name, config.BoolOption, def, desc)
}

func (this Configuration) DurationOption(name string, desc string) Configuration {
	return this.addOption(name, config.DurationOption, time.Duration(0), desc)
}
func (this Configuration) DefaultedDurationOption(name string, def time.Duration, desc string) Configuration {
	return this.addOption(name, config.DurationOption, def, desc)
}

func (this Configuration) addOption(name string, t config.OptionType, def interface{}, desc string) Configuration {
	if this.settings.configs[name] != nil {
		panic(fmt.Sprintf("option %q already defined", name))
	}
	this.settings.configs[name] = controllermanager.NewOptionDefinition(name, t, def, desc)
	return this
}

func (this Configuration) ActivateExplicitly() Configuration {
	this.settings.activateExplicitly = true
	return this
}

func (this Configuration) Definition() Definition {
	return &this.settings
}

func (this Configuration) RegisterAt(registry RegistrationInterface) error {
	return registry.Register(this)
}

func (this Configuration) MustRegisterAt(registry RegistrationInterface) Configuration {
	registry.MustRegister(this)
	return this
}

func (this Configuration) Register() error {
	return registry.Register(this)
}

func (this Configuration) MustRegister() Configuration {
	registry.MustRegister(this)
	return this
}
