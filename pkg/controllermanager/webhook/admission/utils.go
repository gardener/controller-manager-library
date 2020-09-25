/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package admission

import (
	"fmt"

	adminreg "k8s.io/api/admissionregistration/v1beta1"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/extension"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"github.com/gardener/controller-manager-library/pkg/resources/apiextensions"
)

func toResourceSpecs(specs ...extension.ResourceKey) []interface{} {
	result := make([]interface{}, len(specs), len(specs))
	for i, r := range specs {
		result[i] = r
	}
	return result
}

func toClientConfig(cfg *apiextensions.WebhookClientConfig) adminreg.WebhookClientConfig {
	var svc *adminreg.ServiceReference
	if cfg.Service != nil {
		svc = &adminreg.ServiceReference{
			Namespace: cfg.Service.Namespace,
			Name:      cfg.Service.Name,
			Path:      cfg.Service.Path,
			Port:      cfg.Service.PortP(),
		}
	}
	return adminreg.WebhookClientConfig{
		URL:      cfg.URL,
		CABundle: append(cfg.CABundle[:0:0], cfg.CABundle...),
		Service:  svc,
	}
}

func NewAdmissionSpecData(resources resources.ResourcesSource, policy adminreg.FailurePolicyType, ops []adminreg.OperationType, rkeys ...extension.ResourceKey) ([]adminreg.RuleWithOperations, *adminreg.FailurePolicyType, error) {
	var rules []adminreg.RuleWithOperations
	specs := toResourceSpecs(rkeys...)
	for _, spec := range specs {
		rule, err := NewAdmissionRegistration(resources, spec, ops...)
		if err != nil {
			return nil, nil, fmt.Errorf("webhook declaration error: %s", err)
		}
		rules = append(rules, *rule)
	}
	failurePolicy := &policy
	if policy == "" {
		failurePolicy = nil
	}
	return rules, failurePolicy, nil
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
