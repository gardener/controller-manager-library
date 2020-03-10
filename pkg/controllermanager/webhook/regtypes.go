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

package webhook

import (
	"sync"

	"k8s.io/apimachinery/pkg/runtime"
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
