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

package webhooks

import (
	"fmt"

	"github.com/gardener/controller-manager-library/pkg/resources"

	"k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

////////////////////////////////////////////////////////////////////////////////
// WebHookClientSources

type WebhookClientConfigSource interface {
	WebhookClientConfig() *v1beta1.WebhookClientConfig
}

type WebhookClientConfig v1beta1.WebhookClientConfig

func (this *WebhookClientConfig) WebhookClientConfig() *v1beta1.WebhookClientConfig {
	return (*v1beta1.WebhookClientConfig)(this)
}

func NewURLWebhookClientConfig(url string, caBundle []byte) WebhookClientConfigSource {
	return &WebhookClientConfig{
		CABundle: caBundle,
		URL:      &url,
	}
}

func NewDNSWebhookClientConfig(dnsName string, path string, caBundle []byte, port ...int) WebhookClientConfigSource {
	url := fmt.Sprintf("https://%s/%s", dnsName, path)
	if len(port) > 0 && port[0] > 0 {
		url = fmt.Sprintf("https://%s:%d/%s", dnsName, port[0], path)
	}
	return NewURLWebhookClientConfig(url, caBundle)
}

func NewRuntimeServiceWebhookClientConfig(name resources.ObjectName, path string, caBundle []byte, port ...int) WebhookClientConfigSource {
	url := fmt.Sprintf("https://%s.%s/%s", name.Name(), name.Namespace(), path)
	if len(port) > 0 && port[0] > 0 {
		url = fmt.Sprintf("https://%s.%s:%d/%s", name.Name(), name.Namespace(), port[0], path)
	}
	return NewURLWebhookClientConfig(url, caBundle)
}

func NewServiceWebhookClientConfig(name resources.ObjectName, path string, caBundle []byte) WebhookClientConfigSource {
	return &WebhookClientConfig{
		CABundle: caBundle,
		Service: &v1beta1.ServiceReference{
			Namespace: name.Namespace(),
			Name:      name.Name(),
			Path:      &path,
		},
	}
}

////////////////////////////////////////////////////////////////////////////////
// WebHook Specs

type Webhook v1beta1.MutatingWebhook

func toMutating(hooks ...*Webhook) []v1beta1.MutatingWebhook {
	result := make([]v1beta1.MutatingWebhook, len(hooks))
	for i, h := range hooks {
		result[i] = v1beta1.MutatingWebhook(*h)
	}
	return result
}

func NewWebhook(resources resources.ResourcesSource, name string, namespaces *v1.LabelSelector, client WebhookClientConfigSource, specs ...interface{}) (*Webhook, error) {
	var rules []v1beta1.RuleWithOperations
	for _, spec := range specs {
		rule, err := NewAdmissionRegistration(resources, spec)
		if err != nil {
			return nil, err
		}
		rules = append(rules, *rule)
	}

	return &Webhook{
		Name:              name,
		NamespaceSelector: namespaces,
		Rules:             rules,
		ClientConfig:      *client.WebhookClientConfig(),
	}, nil
}

func NewAdmissionRegistration(resources resources.ResourcesSource, spec interface{}, ops ...v1beta1.OperationType) (*v1beta1.RuleWithOperations, error) {
	r, err := resources.Resources().Get(spec)
	if err != nil {
		return nil, err
	}

	if len(ops) == 0 {
		ops = []v1beta1.OperationType{
			v1beta1.Create,
			v1beta1.Update,
		}
	}
	// Create and return RuleWithOperations
	return &v1beta1.RuleWithOperations{
		Operations: ops,
		Rule: v1beta1.Rule{
			APIGroups:   []string{r.GroupVersionKind().Group},
			APIVersions: []string{r.GroupVersionKind().Version},
			Resources:   []string{r.Name()},
		},
	}, nil
}

////////////////////////////////////////////////////////////////////////////////

func CreateOrUpdateMutatingWebhookRegistration(cluster resources.Cluster, name string, webhooks ...*Webhook) error {
	config := &v1beta1.MutatingWebhookConfiguration{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Webhooks: toMutating(webhooks...),
	}
	_, err := cluster.Resources().CreateOrUpdateObject(config)
	return err
}

////////////////////////////////////////////////////////////////////////////////
