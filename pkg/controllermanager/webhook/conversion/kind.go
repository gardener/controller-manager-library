/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package conversion

import (
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/webhook"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"github.com/gardener/controller-manager-library/pkg/resources/apiextensions"
)

func init() {
	webhook.RegisterKindHandlerProvider(webhook.CONVERTING, kindHandlerProvider)
}

type kindHandler struct {
	ext       webhook.Environment
	resources map[schema.GroupKind]webhook.Interface
}

func kindHandlerProvider(ext webhook.Environment, _ webhook.WebhookKind) (webhook.WebhookKindHandler, error) {
	ext.Infof("registering conversion webhook handler")
	h := &kindHandler{ext, map[schema.GroupKind]webhook.Interface{}}
	apiextensions.RegisterClientConfigProvider(h)
	return h, nil
}

func (this *kindHandler) Register(p webhook.Interface) error {
	this.ext.Infof("registering conversion client handlers for %+v", p.GetDefinition().Resources())
	for _, r := range p.GetDefinition().Resources() {
		this.resources[r.GroupKind()] = p
	}
	return nil
}

func (this *kindHandler) GetClientConfig(gk schema.GroupKind, cluster resources.Cluster) apiextensions.WebhookClientConfigSource {
	if p := this.resources[gk]; p != nil {
		cfg, err := this.ext.CreateWebhookClientConfig("using conversion", p.GetDefinition(), cluster)
		if err == nil {
			return cfg
		}
	}
	return nil
}
