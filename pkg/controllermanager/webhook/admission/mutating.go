/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
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
	webhook.RegisterRegistrationHandler(newMutatingHandler())
}

func Mutating(factory AdmissionHandlerType) configuration {
	return configuration{
		settings: _Definition{
			kind:    webhook.MUTATING,
			factory: factory,
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

type MutatingWebhookDeclaration struct {
	adminreg.MutatingWebhook
}

var _ webhook.WebhookDeclaration = (*MutatingWebhookDeclaration)(nil)

func (this *MutatingWebhookDeclaration) Kind() webhook.WebhookKind {
	return webhook.MUTATING
}

func (this *MutatingWebhookDeclaration) DeepCopy() *adminreg.MutatingWebhook {
	return this.MutatingWebhook.DeepCopy()
}

////////////////////////////////////////////////////////////////////////////////

type mutating struct {
	*webhook.RegistrationHandlerBase
}

func newMutatingHandler() *mutating {
	return &mutating{
		webhook.NewRegistrationHandlerBase(webhook.MUTATING, &adminreg.MutatingWebhookConfiguration{}),
	}
}

var _ webhook.RegistrationHandler = (*mutating)(nil)

func (this *mutating) CreateDeclarations(_ logger.LogContext, def webhook.Definition, target cluster.Interface, client apiextensions.WebhookClientConfigSource) (webhook.WebhookDeclarations, error) {
	admindef := def.Handler().(Definition)
	rules, policy, err := NewAdmissionSpecData(target, admindef.GetFailurePolicy(), admindef.GetOperations(), def.Resources()...)
	if err != nil {
		return nil, err
	}
	return webhook.WebhookDeclarations{&MutatingWebhookDeclaration{
		adminreg.MutatingWebhook{
			Name:              def.Name(),
			NamespaceSelector: admindef.GetNamespaces(),
			FailurePolicy:     policy,
			Rules:             rules,
			ClientConfig:      toClientConfig(client.WebhookClientConfig()),
		}},
	}, nil
}

func (this *mutating) Register(ctx webhook.RegistrationContext, labels map[string]string, cluster cluster.Interface, name string, declarations ...webhook.WebhookDeclaration) error {
	config := &adminreg.MutatingWebhookConfiguration{
		ObjectMeta: meta.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Webhooks: toMutating(declarations...),
	}
	var err error
	if len(config.Webhooks) > 0 {
		ctx.Infof("creating mutating webhook %s", name)
		err = resources.FilterObjectDeletionError(cluster.Resources().CreateOrUpdateObject(config))
	}
	return err
}

func (this *mutating) Delete(log logger.LogContext, name string, _ webhook.Definition, cluster cluster.Interface) error {
	r, err := cluster.Resources().Get(&adminreg.MutatingWebhookConfiguration{})
	if err != nil {
		return err
	}
	log.Infof("deleting mutating webhook %s", name)
	return r.DeleteByName(resources.NewObjectName(name))
}

func toMutating(hooks ...webhook.WebhookDeclaration) []adminreg.MutatingWebhook {
	result := make([]adminreg.MutatingWebhook, 0, len(hooks))
	for _, h := range hooks {
		result = append(result, *h.(*MutatingWebhookDeclaration).DeepCopy())
	}
	return result
}
