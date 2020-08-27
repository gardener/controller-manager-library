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

package resources

import (
	"github.com/gardener/controller-manager-library/pkg/utils"
)

type DefaultClusterIdMigration struct {
	migrations map[string]string
}

var _ ClusterIdMigration = &DefaultClusterIdMigration{}
var _ ClusterIdMigrationProvider = &DefaultClusterIdMigration{}

func ClusterIdMigrationFor(clusters ...Cluster) ClusterIdMigration {
	migrations := map[string]string{}
	for _, c := range clusters {
		id := c.GetId()
		for o := range c.GetMigrationIds() {
			migrations[o] = id
		}
	}
	if len(migrations) == 0 {
		return nil
	}
	return &DefaultClusterIdMigration{migrations}
}

func (this *DefaultClusterIdMigration) RequireMigration(id string) string {
	if new, ok := this.migrations[id]; ok {
		return new
	}
	return ""
}

func (this *DefaultClusterIdMigration) GetClusterIdMigration() ClusterIdMigration {
	return this
}

func (this *DefaultClusterIdMigration) String() string {
	m := map[string]utils.StringSet{}

	for k, v := range this.migrations {
		s := m[v]
		if s == nil {
			s = utils.StringSet{}
			m[v] = s
		}
		s.Add(k)
	}

	sep := ""
	r := ""
	for k, v := range m {
		r = r + sep + k + " <- " + v.String()
		sep = ", "
	}
	return r
}
