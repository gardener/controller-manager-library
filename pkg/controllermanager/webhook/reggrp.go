/*
 * Copyright 2020 SAP SE or an SAP affiliate company. All rights reserved.
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
	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	"github.com/gardener/controller-manager-library/pkg/utils"
)

type WebhookRegistrationGroup struct {
	cluster       cluster.Interface
	registrations map[string]utils.StringSet
	declarations  WebhookDeclarations
}

func NewWebhookRegistrationGroup(cluster cluster.Interface) *WebhookRegistrationGroup {
	return &WebhookRegistrationGroup{
		cluster:       cluster,
		registrations: map[string]utils.StringSet{},
		declarations:  WebhookDeclarations{},
	}
}

func (this *WebhookRegistrationGroup) AddDeclaration(d *WebhookDeclaration) {
	this.declarations = append(this.declarations, d)
}

func (this *WebhookRegistrationGroup) AddRegistration(name string, kind WebhookKind) {
	set := this.registrations[name]
	if set == nil {
		set = utils.StringSet{}
		this.registrations[name] = set
	}
	set.Add(string(kind))
}

type WebhookRegistrationGroups map[string]*WebhookRegistrationGroup

func (this WebhookRegistrationGroups) Assure(cluster cluster.Interface) *WebhookRegistrationGroup {
	g := this[cluster.GetName()]
	if g == nil {
		g = NewWebhookRegistrationGroup(cluster)
		this[cluster.GetName()] = g
	}
	return g
}
