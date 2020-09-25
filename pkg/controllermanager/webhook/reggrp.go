/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package webhook

import (
	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	"github.com/gardener/controller-manager-library/pkg/utils"
)

////////////////////////////////////////////////////////////////////////////////

type WebhookRegistrationGroup struct {
	cluster             cluster.Interface
	registrations       map[string]utils.StringSet
	groupedDeclarations map[WebhookKind]WebhookDeclarations
}

func NewWebhookRegistrationGroup(cluster cluster.Interface) *WebhookRegistrationGroup {
	return &WebhookRegistrationGroup{
		cluster:             cluster,
		registrations:       map[string]utils.StringSet{},
		groupedDeclarations: map[WebhookKind]WebhookDeclarations{},
	}
}

func (this *WebhookRegistrationGroup) AddDeclarations(decls ...WebhookDeclaration) {
	for _, d := range decls {
		declarations := this.groupedDeclarations[d.Kind()]
		declarations = append(declarations, d)
		this.groupedDeclarations[d.Kind()] = declarations
	}
}

func (this *WebhookRegistrationGroup) AddRegistrations(kind WebhookKind, names ...string) {
	for _, name := range names {
		set := this.registrations[name]
		if set == nil {
			set = utils.StringSet{}
			this.registrations[name] = set
		}
		set.Add(string(kind))
	}
}

type WebhookRegistrationGroups map[string]*WebhookRegistrationGroup

func (this WebhookRegistrationGroups) GetOrCreateGroup(cluster cluster.Interface) *WebhookRegistrationGroup {
	g := this[cluster.GetId()]
	if g == nil {
		g = NewWebhookRegistrationGroup(cluster)
		this[cluster.GetId()] = g
	}
	return g
}
