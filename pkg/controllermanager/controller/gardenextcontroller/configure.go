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
	"github.com/gardener/controller-manager-library/pkg/utils"
	"time"
)

type Configuration struct {
	config controller.Configuration
}

func Configure(name string) Configuration {
	return Configuration{controller.Configure(name)}.AddFilters(Filter)
}

func (this Configuration) Name(name string) Configuration {
	this.config = this.config.Name(name)
	return this
}

func (this Configuration) MainResource(kind, extensionType string) Configuration {
	this.config = this.config.MainResourceByKey(NewResourceKey(kind, extensionType))
	return this
}

func (this Configuration) MainResourceByKey(key ResourceKey) Configuration {
	this.config = this.config.MainResourceByKey(key)
	return this
}

func (this Configuration) WorkerPool(name string, size int, period time.Duration) Configuration {
	this.config = this.config.WorkerPool(name, size, period)
	return this
}
func (this Configuration) Pool(name string) Configuration {
	this.config = this.config.Pool(name)
	return this
}

func (this Configuration) Cluster(name string) Configuration {
	this.config = this.config.Cluster(name)
	return this
}

func (this Configuration) Watches(keys ...controller.ResourceKey) Configuration {
	this.config = this.config.Watches(keys...)
	return this
}

func (this Configuration) Watch(group, kind string) Configuration {
	this.config = this.config.Watches(controller.NewResourceKey(group, kind))
	return this
}

func (this Configuration) ReconcilerWatches(reconciler string, keys ...controller.ResourceKey) Configuration {
	this.config = this.config.ReconcilerWatches(reconciler, keys...)
	return this
}

func (this Configuration) ReconcilerWatch(reconciler, group, kind string) Configuration {
	this.config = this.config.ReconcilerWatches(reconciler, controller.NewResourceKey(group, kind))
	return this
}

func (this Configuration) Reconciler(t controller.ReconcilerType, name ...string) Configuration {
	this.config = this.config.Reconciler(t, name...)
	return this
}

func (this Configuration) RequireLease() Configuration {
	this.config = this.config.RequireLease()
	return this
}

func (this Configuration) Commands(cmd ...string) Configuration {
	this.config = this.config.Commands(cmd...)
	return this
}
func (this Configuration) CommandMatchers(cmd ...utils.Matcher) Configuration {
	this.config = this.config.CommandMatchers(cmd...)
	return this
}

func (this Configuration) ReconcilerCommands(reconciler string, cmd ...string) Configuration {
	this.config = this.config.ReconcilerCommands(reconciler, cmd...)
	return this
}
func (this Configuration) ReconcilerCommandMatchers(reconciler string, cmd ...utils.Matcher) Configuration {
	this.config = this.config.ReconcilerCommandMatchers(reconciler, cmd...)
	return this
}

func (this Configuration) AddFilters(f ...controller.ResourceFilter) Configuration {
	this.config = this.config.AddFilters(f...)
	return this
}
func (this Configuration) Filters(f ...controller.ResourceFilter) Configuration {
	this.config = this.config.Filters(f...).AddFilters(Filter)
	return this
}

func (this Configuration) Definition() controller.Definition {
	return this.config.Definition()
}
func (this Configuration) Register(group ...string) error {
	return Register(this.Definition(), group...)
}
func (this Configuration) MustRegister(group ...string) {
	MustRegister(this.Definition(), group...)
}
