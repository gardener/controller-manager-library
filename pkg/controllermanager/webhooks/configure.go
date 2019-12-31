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

package webhooks

import (
	"fmt"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type _Definition struct {
	name               string
	keys               []controller.ResourceKey
	cluster            string
	kind               string
	handler            AdmissionHandlerType
	namespaces         *metav1.LabelSelector
	activateExplicitly bool
}

var _ Definition = &_Definition{}

func (this *_Definition) GetResources() []controller.ResourceKey {
	return append(this.keys[:0:0], this.keys...)
}
func (this *_Definition) GetName() string {
	return this.name
}
func (this *_Definition) GetCluster() string {
	return this.cluster
}
func (this *_Definition) GetKind() string {
	return this.kind
}
func (this *_Definition) GetHandlerType() AdmissionHandlerType {
	return this.handler
}
func (this *_Definition) GetNamespaces() *metav1.LabelSelector {
	return this.namespaces
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
			name: name,
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

func (this Configuration) Resource(group, kind string) Configuration {
	this.settings.keys = append(this.settings.keys, controller.NewResourceKey(group, kind))
	return this
}

func (this Configuration) Namespaces(selector *metav1.LabelSelector) Configuration {
	this.settings.namespaces = selector
	return this
}

func (this Configuration) Handler(htype AdmissionHandlerType) Configuration {
	this.settings.handler = htype
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
