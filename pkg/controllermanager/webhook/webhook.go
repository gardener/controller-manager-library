/*
 * Copyright 2020 SAP SE or an SAP affiliate company. All rights reserved.
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

package webhook

import (
	"github.com/gardener/controller-manager-library/pkg/controllermanager"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/webhook/admission"
	"github.com/gardener/controller-manager-library/pkg/resources"
)

type webhook struct {
	controllermanager.ElementBase

	config      *WebhookConfig
	extension   *Extension
	definition  Definition
	cluster     cluster.Interface
	hook        admission.Interface
	httphandler *admission.HTTPHandler
}

var _ Interface = &webhook{}

func NewWebhook(ext *Extension, def Definition, cluster cluster.Interface) (*webhook, error) {
	options := ext.GetConfig().GetSource(def.GetName()).(*WebhookConfig)
	this := &webhook{
		extension:  ext,
		definition: def,
		config:     options,
		cluster:    cluster,
	}
	this.ElementBase = controllermanager.NewElementBase(ext.GetContext(), ctx_webhook, this, def.GetName(), options)
	hook, err := def.GetHandlerType()(this)
	if err != nil {
		return nil, err
	}
	this.hook = hook
	this.httphandler = admission.New(this, this.cluster.ResourceContext().Scheme(), this.hook)
	return this, nil
}

func (this *webhook) GetResources() resources.Resources {
	return this.cluster.Resources()
}

func (this *webhook) GetEnvironment() Environment {
	return this.extension
}

func (this *webhook) GetDefinition() Definition {
	return this.definition
}

func (this *webhook) GetHTTPHandler() *admission.HTTPHandler {
	return this.httphandler
}
