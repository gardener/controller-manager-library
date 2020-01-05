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

package webhook

import (
	"fmt"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"github.com/gardener/controller-manager-library/pkg/server"

	adminreg "k8s.io/api/admissionregistration/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

////////////////////////////////////////////////////////////////////////////////
// WebHookClientSources

type WebhookClientConfigSource interface {
	WebhookClientConfig() *adminreg.WebhookClientConfig
}

type WebhookClientConfig adminreg.WebhookClientConfig

func (this *WebhookClientConfig) WebhookClientConfig() *adminreg.WebhookClientConfig {
	return (*adminreg.WebhookClientConfig)(this)
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
	path = server.NormPath(path)
	return &WebhookClientConfig{
		CABundle: caBundle,
		Service: &adminreg.ServiceReference{
			Namespace: name.Namespace(),
			Name:      name.Name(),
			Path:      &path,
		},
	}
}

////////////////////////////////////////////////////////////////////////////////
// WebHook Specs

type WebhookDeclaration struct {
	adminreg.MutatingWebhook
	Kind WebhookKind
}

type WebhookDeclarations []*WebhookDeclaration

func toMutating(hooks ...*WebhookDeclaration) []adminreg.MutatingWebhook {
	result := make([]adminreg.MutatingWebhook, 0, len(hooks))
	for _, h := range hooks {
		if h.Kind == MUTATING {
			result = append(result, h.MutatingWebhook)
		}
	}
	return result
}

func toValidating(hooks ...*WebhookDeclaration) []adminreg.ValidatingWebhook {
	result := make([]adminreg.ValidatingWebhook, 0, len(hooks))
	for _, h := range hooks {
		if h.Kind == VALIDATING {
			hook := adminreg.ValidatingWebhook{
				Name:                    h.Name,
				ClientConfig:            h.ClientConfig,
				Rules:                   h.Rules,
				FailurePolicy:           h.FailurePolicy,
				MatchPolicy:             h.MatchPolicy,
				NamespaceSelector:       h.NamespaceSelector,
				ObjectSelector:          h.ObjectSelector,
				SideEffects:             h.SideEffects,
				TimeoutSeconds:          h.TimeoutSeconds,
				AdmissionReviewVersions: h.AdmissionReviewVersions,
			}
			result = append(result, hook)
		}
	}
	return result
}

func NewWebhookDeclaration(kind WebhookKind, resources resources.ResourcesSource, name string, namespaces *meta.LabelSelector, policy adminreg.FailurePolicyType, client WebhookClientConfigSource, ops []adminreg.OperationType, specs ...interface{}) (*WebhookDeclaration, error) {
	var rules []adminreg.RuleWithOperations
	for _, spec := range specs {
		rule, err := NewAdmissionRegistration(resources, spec, ops...)
		if err != nil {
			return nil, fmt.Errorf("webhook declaration error: %s", err)
		}
		rules = append(rules, *rule)
	}

	failurePolicy := &policy
	if policy == "" {
		failurePolicy = nil
	}

	return &WebhookDeclaration{
		Kind: kind,
		MutatingWebhook: adminreg.MutatingWebhook{
			Name:              name,
			NamespaceSelector: namespaces,
			FailurePolicy:     failurePolicy,
			Rules:             rules,
			ClientConfig:      *client.WebhookClientConfig(),
		},
	}, nil
}

func NewAdmissionRegistration(resources resources.ResourcesSource, spec interface{}, ops ...adminreg.OperationType) (*adminreg.RuleWithOperations, error) {
	r, err := resources.Resources().Get(spec)
	if err != nil {
		return nil, fmt.Errorf("admission registration error: %s", err)
	}

	if len(ops) == 0 {
		ops = []adminreg.OperationType{
			adminreg.Create,
			adminreg.Update,
		}
	}
	// Create and return RuleWithOperations
	return &adminreg.RuleWithOperations{
		Operations: ops,
		Rule: adminreg.Rule{
			APIGroups:   []string{r.GroupVersionKind().Group},
			APIVersions: []string{r.GroupVersionKind().Version},
			Resources:   []string{r.Name()},
		},
	}, nil
}

////////////////////////////////////////////////////////////////////////////////

func AddLabel(labels map[string]string, key, value string) map[string]string {
	new := map[string]string{}
	for k, v := range labels {
		new[k] = v
	}
	new[key] = value
	return new
}

func CreateOrUpdateMutatingWebhookRegistration(labels map[string]string, cluster resources.Cluster, name string, webhooks ...*WebhookDeclaration) (int, error) {
	config := &adminreg.MutatingWebhookConfiguration{
		ObjectMeta: meta.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Webhooks: toMutating(webhooks...),
	}
	var err error
	if len(config.Webhooks) > 0 {
		_, err = cluster.Resources().CreateOrUpdateObject(config)
	}
	return len(config.Webhooks), err
}

func CreateOrUpdateValidatingWebhookRegistration(labels map[string]string, cluster resources.Cluster, name string, webhooks ...*WebhookDeclaration) (int, error) {
	config := &adminreg.ValidatingWebhookConfiguration{
		ObjectMeta: meta.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Webhooks: toValidating(webhooks...),
	}
	var err error
	if len(config.Webhooks) > 0 {
		_, err = cluster.Resources().CreateOrUpdateObject(config)
	}
	return len(config.Webhooks), err
}

////////////////////////////////////////////////////////////////////////////////

func DeleteMutatingWebhookRegistration(cluster resources.ResourcesSource, name string) error {
	r,err:=cluster.Resources().Get(&adminreg.MutatingWebhookConfiguration{})
	if err != nil {
	    return err
	}
	return r.DeleteByName(resources.NewObjectName(name))
}

func DeleteValidatingWebhookRegistration(cluster resources.ResourcesSource, name string) error {
	r,err:=cluster.Resources().Get(&adminreg.ValidatingWebhookConfiguration{})
	if err != nil {
		return err
	}
	return r.DeleteByName(resources.NewObjectName(name))
}