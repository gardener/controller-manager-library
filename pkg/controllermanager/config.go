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

package controllermanager

import (
	"time"

	"github.com/gardener/controller-manager-library/pkg/config"
)

var GracePeriod time.Duration

const OPTION_SOURCE = "controllermanager"

type Config struct {
	Controllers                 string
	Name                        string
	OmitLease                   bool
	DisableNamespaceRestriction bool
	NamespaceRestriction        bool
}

var _ config.OptionSource = (*Config)(nil)

func (this *Config) AddOptionsToSet(set config.OptionSet) {
	set.AddDurationOption(&GracePeriod, "grace-period", "", 0, "inactivity grace period for detecting end of cleanup for shutdown")
	set.AddStringOption(&this.Name, "name", "", "", "name used for controller manager")
	set.AddBoolOption(&this.OmitLease, "omit-lease", "", false, "omit lease for development")
	set.AddStringOption(&this.Controllers, "controllers", "c", "all", "comma separated list of controllers to start (<name>,source,target,all)")
	set.AddBoolOption(&this.NamespaceRestriction, "namespace-local-access-only", "n", false, "enable access restriction for namespace local access only (deprecated)")
	set.AddBoolOption(&this.DisableNamespaceRestriction, "disable-namespace-restriction", "", false, "disable access restriction for namespace local access only")
}
