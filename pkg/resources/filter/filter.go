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
 *
 */

package filter

import (
	"k8s.io/apimachinery/pkg/runtime/schema"

	. "github.com/gardener/controller-manager-library/pkg/resources"
)

/////////////////////////////////////////////////////////////////////////////////

func AllObjects(key ClusterObjectKey) bool {
	return true
}

func NoObjects(key ClusterObjectKey) bool {
	return false
}

func GroupKindFilter(gk ...schema.GroupKind) KeyFilter {
	return GroupKindFilterBySet(NewGroupKindSet(gk...))
}

func GroupKindFilterBySet(gks GroupKindSet) KeyFilter {
	return func(key ClusterObjectKey) bool {
		return gks.Contains(key.GroupKind())
	}
}

func ClusterGroupKindFilter(cgk ...ClusterGroupKind) KeyFilter {
	return ClusterGroupKindFilterBySet(NewClusterGroupKindSet(cgk...))
}

func ClusterGroupKindFilterBySet(gks ClusterGroupKindSet) KeyFilter {
	return func(key ClusterObjectKey) bool {
		return gks.Contains(key.ClusterGroupKind())
	}
}

func Or(filters ...KeyFilter) KeyFilter {
	return func(key ClusterObjectKey) bool {
		for _, f := range filters {
			if f(key) {
				return true
			}
		}
		return false
	}
}

func And(filters ...KeyFilter) KeyFilter {
	return func(key ClusterObjectKey) bool {
		if len(filters) == 0 {
			return false
		}
		for _, f := range filters {
			if !f(key) {
				return false
			}
		}
		return true
	}
}
