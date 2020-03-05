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

package apiextensions

import (
	"github.com/Masterminds/semver"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/gardener/controller-manager-library/pkg/resources"
	"github.com/gardener/controller-manager-library/pkg/utils"
)

type CRDVersion string

const CRD_V1 = CRDVersion("v1")
const CRD_V1BETA1 = CRDVersion("v1beta1")

type CustomResourceDefinitionVersions struct {
	versioned *utils.Versioned
}

var v116 = semver.MustParse("1.16.0")
var otype runtime.Object

func NewCustomResourceDefinitionVersions(crd ...interface{}) (*CustomResourceDefinitionVersions, error) {
	if len(crd) > 1 {
		return nil, nil
	}
	def := &CustomResourceDefinitionVersions{utils.NewVersioned(&CustomResourceDefinition{})}
	if len(crd) > 0 {
		o, err := CRDObject(crd)
		if err != nil {
			return nil, err
		}
		def.versioned.SetDefault(o)
	}
	return def, nil
}

func (this *CustomResourceDefinitionVersions) GetFor(cluster resources.Cluster) runtime.Object {
	f := this.versioned.GetFor(cluster.GetServerVersion())
	if f != nil {
		crd := f.(*CustomResourceDefinition)
		return crd.For(cluster)
	}
	return nil
}

func (this *CustomResourceDefinitionVersions) RegisterVersion(v *semver.Version, spec CRDSpecification) *CustomResourceDefinitionVersions {
	crd, err := CRDObject(spec)
	utils.Must(err)
	this.versioned.MustRegisterVersion(v, crd)
	return this
}

func toClientConfig(cfg *WebhookClientConfig) *apiextensions.WebhookClientConfig {
	var svc *apiextensions.ServiceReference
	if cfg.Service != nil {
		svc = &apiextensions.ServiceReference{
			Namespace: cfg.Service.Namespace,
			Name:      cfg.Service.Name,
			Path:      cfg.Service.Path,
			Port:      cfg.Service.Port,
		}
	}
	return &apiextensions.WebhookClientConfig{
		URL:      cfg.URL,
		CABundle: append(cfg.CABundle[:0:0], cfg.CABundle...),
		Service:  svc,
	}
}
