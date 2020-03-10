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

package admission

import (
	adminreg "k8s.io/api/admissionregistration/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/webhook"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"github.com/gardener/controller-manager-library/pkg/resources/apiextensions"
)

func init() {
	webhook.RegisterRegistrationHandler(newValidatingHandler())
}

func Validating(factory AdmissionHandlerType) configuration {
	return configuration{
		settings: _Definition{
			kind:    webhook.VALIDATING,
			factory: factory,
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

type ValidatingWebhookDeclaration struct {
	adminreg.ValidatingWebhook
}

var _ webhook.WebhookDeclaration = (*ValidatingWebhookDeclaration)(nil)

func (this *ValidatingWebhookDeclaration) Kind() webhook.WebhookKind {
	return webhook.VALIDATING
}

func (this *ValidatingWebhookDeclaration) DeepCopy() *adminreg.ValidatingWebhook {
	return this.ValidatingWebhook.DeepCopy()
}

////////////////////////////////////////////////////////////////////////////////

type validating struct {
	*webhook.RegistrationHandlerBase
}

func newValidatingHandler() *validating {
	return &validating{
		webhook.NewRegistrationHandlerBase(webhook.VALIDATING, &adminreg.ValidatingWebhookConfiguration{}),
	}
}

var _ webhook.RegistrationHandler = (*mutating)(nil)

func (this *validating) CreateDeclarations(log logger.LogContext, def webhook.Definition, target cluster.Interface, client apiextensions.WebhookClientConfigSource) (webhook.WebhookDeclarations, error) {
	admindef := def.Handler().(Definition)
	rules, policy, err := NewAdmissionSpecData(target, admindef.GetFailurePolicy(), admindef.GetOperations(), def.Resources()...)
	if err != nil {
		return nil, err
	}
	return webhook.WebhookDeclarations{&ValidatingWebhookDeclaration{
		adminreg.ValidatingWebhook{
			Name:              def.Name(),
			NamespaceSelector: admindef.GetNamespaces(),
			FailurePolicy:     policy,
			Rules:             rules,
			ClientConfig:      toClientConfig(client.WebhookClientConfig()),
		}},
	}, nil
}

func (this *validating) Register(ctx webhook.RegistrationContext, labels map[string]string, cluster cluster.Interface, name string, declarations ...webhook.WebhookDeclaration) error {
	config := &adminreg.ValidatingWebhookConfiguration{
		ObjectMeta: meta.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Webhooks: toValidating(declarations...),
	}
	var err error
	if len(config.Webhooks) > 0 {
		ctx.Infof("deleting validating webhook %s", name)
		err = resources.FilterObjectDeletionError(cluster.Resources().CreateOrUpdateObject(config))
	}
	return err
}

func (this *validating) Delete(log logger.LogContext, name string, def webhook.Definition, cluster cluster.Interface) error {
	r, err := cluster.Resources().Get(&adminreg.ValidatingWebhookConfiguration{})
	if err != nil {
		return err
	}
	log.Infof("deleting validating webhook %s", name)
	return r.DeleteByName(resources.NewObjectName(name))
}

func toValidating(hooks ...webhook.WebhookDeclaration) []adminreg.ValidatingWebhook {
	result := make([]adminreg.ValidatingWebhook, 0, len(hooks))
	for _, h := range hooks {
		result = append(result, *h.(*ValidatingWebhookDeclaration).DeepCopy())
	}
	return result
}
