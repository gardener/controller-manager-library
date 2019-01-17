/*
 * Copyright 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package gardenextcontroller

import (
	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"github.com/gardener/controller-manager-library/pkg/resources/gardenextensions"
)

func Register(reg controller.Registerable, group ...string) error {
	return controller.Register(controller.AddFilters(reg.Definition(), Filter), group...)
}

func MustRegister(reg controller.Registerable, group ...string) {
	controller.MustRegister(controller.AddFilters(reg.Definition(), Filter), group...)
}

func Filter(owning controller.ResourceKey, obj resources.Object) bool {
	switch k := owning.(type) {
	case ResourceKey:
		extension, ok := obj.Data().(gardenextensions.ExtensionInterface)
		if ok && obj.GroupKind() == k.GroupKind() && extension.GetExtensionType() != k.ExtensionType() {
			return false
		}
		return true
	}
	return controller.Filter(owning, obj)
}
