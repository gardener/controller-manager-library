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

package controller

import (
	"fmt"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/utils"
)

////////////////////////////////////////////////////////////////////////////////

type controllers []*controller

func (this controllers) Contains(cntr *controller) bool {
	for _, c := range this {
		if c == cntr {
			return true
		}
	}
	return false
}

func (this controllers) Get(name string) *controller {
	for _, c := range this {
		if c.GetName() == name {
			return c
		}
	}
	return nil
}

func (this controllers) getOrder(logger logger.LogContext) (controllers, error) {
	order := controllers{}
	stack := controllers{}
	for _, c := range this {
		err := this._orderAdd(logger, &order, stack, c)
		if err != nil {
			return nil, err
		}
	}
	return order, nil
}

func (this controllers) _orderAdd(logger logger.LogContext, order *controllers, stack controllers, c *controller) error {
	if stack.Contains(c) {
		cycle := ""
		for _, s := range stack {
			if cycle != "" || s == c {
				if cycle != "" {
					cycle += " -> "
				}
				cycle += c.GetName()
			}
		}
		return fmt.Errorf("controller startup cycle detected: %s -> %s", cycle, c.GetName())
	}
	if !(*order).Contains(c) {
		stack = append(stack, c)
		after := c.GetDefinition().After()
		if len(after) > 0 {
			preferred := []string{}
			for _, a := range after {
				if dep := this.Get(a); dep != nil {
					preferred = append(preferred, a)
					err := this._orderAdd(logger, order, stack, dep)
					if err != nil {
						return err
					}
				}
			}
			if len(preferred) > 0 {
				logger.Infof("  %s needs to be started after %s", utils.Strings(preferred...))
			}
		}
		*order = append(*order, c)
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type StartupGroup interface {
	Startup() error
	Add(c *controller)
	Controllers() controllers
}

type startupgroup struct {
	extension   *Extension
	cluster     cluster.Interface
	controllers controllers
}

func (this *startupgroup) Add(c *controller) {
	this.controllers = append(this.controllers, c)
}

func (this *startupgroup) Startup() error {
	for _, c := range this.controllers {
		err := this.extension.startController(c)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *startupgroup) Controllers() controllers {
	return this.controllers
}

////////////////////////////////////////////////////////////////////////////////

func (this *Extension) getPlainStartupGroup(cluster cluster.Interface) StartupGroup {
	g := this.plain_groups[cluster.GetName()]
	if g == nil {
		g = &startupgroup{this, cluster, nil}
		this.plain_groups[cluster.GetName()] = g
	}
	return g
}

func (this *Extension) getLeaseStartupGroup(cluster cluster.Interface) StartupGroup {
	g := this.lease_groups[cluster.GetName()]
	if g == nil {
		g = &leasestartupgroup{startupgroup{this, cluster, nil}}
		this.lease_groups[cluster.GetName()] = g
	}
	return g
}

func (this *Extension) startGroups(grps ...map[string]StartupGroup) error {
	for _, grp := range grps {
		for _, g := range grp {
			err := g.Startup()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
