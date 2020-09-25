/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package webhook

import (
	"sync"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/extension"
)

var lock sync.Mutex
var registrationtypes = map[WebhookKind]RegistrationHandler{}

type RegistrationResources map[WebhookKind]runtime.Object

func RegisterRegistrationHandler(r RegistrationHandler) {
	lock.Lock()
	defer lock.Unlock()
	registrationtypes[r.Kind()] = r
}

func GetRegistrationHandler(kind WebhookKind) RegistrationHandler {
	lock.Lock()
	defer lock.Unlock()
	return registrationtypes[kind]
}

func GetRegistrationResources() RegistrationResources {
	lock.Lock()
	defer lock.Unlock()

	resources := RegistrationResources{}
	for _, r := range registrationtypes {
		o := r.RegistrationResource()
		if o != nil {
			resources[r.Kind()] = o
		}
	}
	return resources
}

////////////////////////////////////////////////////////////////////////////////

type RegistrationHandlerBase struct {
	kind  WebhookKind
	proto runtime.Object
}

func NewRegistrationHandlerBase(kind WebhookKind, obj runtime.Object) *RegistrationHandlerBase {
	return &RegistrationHandlerBase{kind, obj}
}

func (this *RegistrationHandlerBase) OptionSourceCreator() extension.OptionSourceCreator {
	return nil
}

func (this *RegistrationHandlerBase) Kind() WebhookKind {
	return this.kind
}

func (this *RegistrationHandlerBase) RegistrationResource() runtime.Object {
	return this.proto
}

func (this *RegistrationHandlerBase) RequireDedicatedRegistrations() bool {
	return false
}

func (this *RegistrationHandlerBase) RegistrationNames(def Definition) []string {
	return []string{def.Name()}
}
