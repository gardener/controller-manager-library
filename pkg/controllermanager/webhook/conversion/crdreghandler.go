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

package conversion

import (
	adminreg "k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/webhook"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"github.com/gardener/controller-manager-library/pkg/resources/apiextensions"
)

func init() {
	webhook.RegisterRegistrationHandler(newCRDHandler())
}

////////////////////////////////////////////////////////////////////////////////

type CRDDeclaration struct {
	resources.ObjectData
}

var _ webhook.WebhookDeclaration = (*CRDDeclaration)(nil)

func (this *CRDDeclaration) Kind() webhook.WebhookKind {
	return webhook.CONVERTING
}

////////////////////////////////////////////////////////////////////////////////

type crdhandler struct {
	*webhook.RegistrationHandlerBase
}

func newCRDHandler() *crdhandler {
	return &crdhandler{
		webhook.NewRegistrationHandlerBase(webhook.CONVERTING, &v1beta1.CustomResourceDefinition{}),
	}
}

var _ webhook.RegistrationHandler = (*crdhandler)(nil)

func (this *crdhandler) RequireDedicatedRegistrations() bool {
	return true
}

func (this *crdhandler) RegistrationNames(def webhook.Definition) []string {
	names := []string{}
	for _, r := range def.GetResources() {
		crd := apiextensions.GetCRD(r.GroupKind())
		if crd == nil {
			continue
		}
		names = append(names, crd.Name)
	}
	return names
}

func (this *crdhandler) CreateDeclarations(log logger.LogContext, def webhook.Definition, target cluster.Interface, client apiextensions.WebhookClientConfigSource) (webhook.WebhookDeclarations, error) {
	result := webhook.WebhookDeclarations{}
	log.Infof("creating crd manifests of %s(%s) for cluster %s(%s)", def.GetName(), def.GetKind(), target.GetId(), target.GetServerVersion())
	for _, r := range def.GetResources() {
		crd := apiextensions.GetCRD(r.GroupKind())
		if crd == nil {
			log.Infof("  no crd for %s", r.GroupKind())
			continue
		}
		o := crd.For(target)
		if o == nil {
			log.Infof("  %s not available for cluster", r.GroupKind())
			continue
		}
		log.Infof("  %s", o.GetName())
		result = append(result, &CRDDeclaration{o})
	}
	return result, nil
}

func (this *crdhandler) Register(log logger.LogContext, labels map[string]string, cluster cluster.Interface, name string, declarations ...webhook.WebhookDeclaration) error {
	log.Infof("registering crds...")
	for _, d := range declarations {
		decl := d.(*CRDDeclaration).ObjectData
		decl.SetLabels(labels)
		log.Infof("   %s (%T)", decl.GetName(), decl)
		err := apiextensions.CreateCRDFromObject(cluster, decl)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *crdhandler) Delete(name string, cluster cluster.Interface) error {
	r, err := cluster.Resources().Get(&adminreg.MutatingWebhookConfiguration{})
	if err != nil {
		return err
	}
	return r.DeleteByName(resources.NewObjectName(name))
}
