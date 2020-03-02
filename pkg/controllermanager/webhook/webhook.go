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
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/extension"
	"github.com/gardener/controller-manager-library/pkg/resources"
)

type webhook struct {
	extension.ElementBase

	config     *WebhookConfig
	extension  *Extension
	definition Definition
	scheme     *runtime.Scheme
	cluster    cluster.Interface
}

var _ Interface = &webhook{}

func NewWebhook(ext *Extension, def Definition, cluster cluster.Interface) (*webhook, error) {
	var err error

	scheme := def.GetScheme()
	options := ext.GetConfig().GetSource(def.GetName()).(*WebhookConfig)
	if scheme != nil && cluster != nil {
		cluster, err = ext.GetClusters().Cache().WithScheme(cluster, scheme)
		if err != nil {
			return nil, err
		}
	} else {
		scheme = cluster.ResourceContext().Scheme()
	}
	this := &webhook{
		extension:  ext,
		definition: def,
		config:     options,
		cluster:    cluster,
		scheme:     scheme,
	}
	this.ElementBase = extension.NewElementBase(ext.GetContext(), ctx_webhook, this, def.GetName(), options)
	if err != nil {
		return nil, err
	}
	return this, nil
}

func (this *webhook) GetResources() resources.Resources {
	if this.cluster == nil {
		return nil
	}
	return this.cluster.Resources()
}

func (this *webhook) GetEnvironment() Environment {
	return this.extension
}

func (this *webhook) GetDefinition() Definition {
	return this.definition
}

func (this *webhook) GetCluster() cluster.Interface {
	return this.cluster
}

func (this *webhook) GetScheme() *runtime.Scheme {
	return this.scheme
}
