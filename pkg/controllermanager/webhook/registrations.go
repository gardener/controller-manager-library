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
	"fmt"
	"sync"
	"time"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/utils"
	"github.com/gardener/controller-manager-library/pkg/wait"
)

////////////////////////////////////////////////////////////////////////////////

type MaintainedRegistration struct {
	name         string
	def          Definition
	handler      RegistrationHandler
	cluster      cluster.Interface
	declarations WebhookDeclarations
}

func (this *MaintainedRegistration) register(ext *Extension) error {
	return this.handler.Register(ext, ext.labels, this.cluster, this.name, this.declarations...)
}

type MaintainedRegistrations struct {
	lock          sync.Mutex
	registrations []*MaintainedRegistration
	pending       []*MaintainedRegistration
	next          int
}

func (this *MaintainedRegistrations) addRegistration(handler RegistrationHandler, name string, def Definition, cluster cluster.Interface, declarations ...WebhookDeclaration) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.registrations = append(this.registrations, &MaintainedRegistration{
		name:         name,
		def:          def,
		handler:      handler,
		cluster:      cluster,
		declarations: declarations,
	})
}

func (this *MaintainedRegistrations) removeRegistration(handler RegistrationHandler, name string, def Definition, cluster cluster.Interface) {
	this.lock.Lock()
	defer this.lock.Unlock()

	for i, m := range this.registrations {
		if m.handler != handler {
			continue
		}
		if m.cluster != cluster {
			continue
		}
		if m.def != def {
			continue
		}
		if m.name != name {
			continue
		}
		this.registrations = append(this.registrations[:i], this.registrations[i+1:]...)
		for i, p := range this.pending {
			if p == m {
				this.registrations = append(this.registrations[:i], this.registrations[i+1:]...)
				break
			}
		}
		break
	}
}

func (this *MaintainedRegistrations) TriggerRegistrationUPdate(ext *Extension) {
	this.lock.Lock()
	defer this.lock.Unlock()

	if len(this.registrations) > 0 {
		start := this.pending == nil
		this.pending = append(this.registrations[:0:0], this.registrations...)
		this.next = 0
		ext.Info("update %d registrations", len(this.pending))
		if start {
			go func() {
				ext.Info("starting registration update handler")
				backoff := wait.Backoff{
					Steps:    -1,
					Duration: 1 * time.Second,
					Factor:   1.1,
					Cap:      10 * time.Minute,
				}
				err := wait.ExponentialBackoff(ext.GetContext(), backoff, func() (bool, error) {
					return this._driveRegistrations(ext), nil
				})
				if err != nil {
					ext.Info("registration update cancelled: %s", err)
				} else {
					ext.Info("registration update done")
				}
			}()
		}
	}
}

func (this *MaintainedRegistrations) _next() *MaintainedRegistration {
	this.lock.Lock()
	defer this.lock.Unlock()

	if this.next >= len(this.pending) {
		this.pending = nil
		return nil
	}
	this.next++
	return this.pending[this.next-1]
}

func (this *MaintainedRegistrations) _driveRegistrations(ext *Extension) bool {
	failed := false
	for {
		m := this._next()
		if m == nil {
			return !failed
		}
		err := m.register(ext)
		if err != nil {
			ext.Errorf("error during registration update: %s", err)
			failed = true
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

func (this *Extension) addRegistration(handler RegistrationHandler, name string, def Definition, cluster cluster.Interface, declarations ...WebhookDeclaration) error {
	this.maintained.addRegistration(handler, name, def, cluster, declarations...)
	return handler.Register(this, this.labels, cluster, name, declarations...)
}

func (this *Extension) removeRegistration(handler RegistrationHandler, name string, def Definition, cluster cluster.Interface) error {
	this.maintained.removeRegistration(handler, name, def, cluster)
	return handler.Delete(this, name, def, cluster)
}

func (this *Extension) handleRegistrationGroups(regs WebhookRegistrationGroups, name string, cleanup bool) error {
	for n, g := range regs {
		if len(g.groupedDeclarations) > 0 {
			msg := logger.NewOptionalSingletonMessage(this.Infof, "registering grouped webhooks for cluster %q with name %s", n, name)
			for k, declarations := range g.groupedDeclarations {
				handler := GetRegistrationHandler(k)
				if handler != nil && len(declarations) > 0 {
					msg.Once()
					err := this.addRegistration(handler, name, nil, g.cluster, declarations...)
					if err != nil {
						return err
					}
					g.AddRegistrations(k, name)
					this.Infof("  found %d %s webhooks for cluster %q", len(declarations), k, n)
				}
			}
		}
		selector := labels.NewSelector()
		for k, v := range this.labels {
			r, err := labels.NewRequirement(k, selection.Equals, []string{v})
			if err != nil {
				return err
			}
			selector = selector.Add(*r)
		}

		if cleanup {
			this.Infof("looking for obsolete registrations: %s", selector.String())
			err := this.cleanup(g.cluster, selector, g.registrations, GetRegistrationResources())
			if err != nil {
				return err
			}
		}
		this.Infof("registrations done")
	}
	return nil
}

func (this *Extension) cleanup(cluster cluster.Interface, selector labels.Selector, keep map[string]utils.StringSet, examples RegistrationResources) error {
	for key, example := range examples {
		r, err := cluster.Resources().GetByExample(example)
		if err != nil {
			return err
		}
		kind := r.Info().Kind()
		if err != nil {
			return err
		}

		list, err := r.List(meta.ListOptions{LabelSelector: selector.String()})
		if err != nil {
			return err
		}

		this.Infof("found %d matching %ss  (%s)", len(list), kind, key)
		for _, found := range list {
			if !keep[found.GetName()].Contains(string(key)) {
				this.Infof("found obsolete %s %q (%s) in cluster %q", kind, found.GetName(), keep[found.GetName()], cluster.GetName())
				err := found.Delete()
				if err != nil {
					return fmt.Errorf("cannot delete obsolete %s %q in cluster %q: %s", kind, found.GetName(), cluster.GetName(), err)
				}
			}
		}
	}
	return nil
}
