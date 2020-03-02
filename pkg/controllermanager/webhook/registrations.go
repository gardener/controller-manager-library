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

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/utils"
)

////////////////////////////////////////////////////////////////////////////////

func (this *Extension) register(regs WebhookRegistrationGroups, name string, cleanup bool) error {
	for n, g := range regs {
		if len(g.declarations) > 0 {
			msg := logger.NewOptionalSingletonMessage(this.Infof, "registering webhooks for cluster %q (%s)", n, name)
			for k, declarations := range g.declarations {
				handler := GetRegistrationHandler(k)
				if handler != nil && len(declarations) > 0 {
					msg.Once()
					err := handler.Register(this, this.labels, g.cluster, name, declarations...)
					if err != nil {
						return err
					}
					g.AddRegistration(name, k)
					this.Infof("  found %d %s webhooks", len(declarations), k)
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
		this.Infof("looking for obsolete registrations: %s", selector.String())

		if cleanup {
			err := this.cleanup(g.cluster, selector, g.registrations, RegistrationResources()...)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (this *Extension) cleanup(cluster cluster.Interface, selector labels.Selector, keep map[string]utils.StringSet, examples ...runtime.Object) error {
	for _, example := range examples {
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

		key := string(kinds[kind])

		this.Infof("found %d matching %ss  (%s)", len(list), kind, key)
		for _, found := range list {
			if !keep[found.GetName()].Contains(key) {
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
