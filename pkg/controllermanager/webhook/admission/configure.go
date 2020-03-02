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
	"fmt"
	"net/http"

	adminreg "k8s.io/api/admissionregistration/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/webhook"
)

type Definition interface {
	GetKind() webhook.WebhookKind
	GetHTTPHandler(wh webhook.Interface) (http.Handler, error)

	GetNamespaces() *meta.LabelSelector
	GetOperations() []adminreg.OperationType
	GetFailurePolicy() adminreg.FailurePolicyType
}

type _Definition struct {
	kind       webhook.WebhookKind
	factory    AdmissionHandlerType
	namespaces *meta.LabelSelector
	operations []adminreg.OperationType
	policy     adminreg.FailurePolicyType
}

var _ webhook.WebhookHandler = (*_Definition)(nil)
var _ Definition = (*_Definition)(nil)

func (this *_Definition) GetKind() webhook.WebhookKind {
	return this.kind
}

func (this *_Definition) GetHTTPHandler(wh webhook.Interface) (http.Handler, error) {
	h, err := this.factory(wh)
	if err != nil {
		return nil, err
	}
	return &HTTPHandler{webhook: h, LogContext: wh}, nil
}

func (this *_Definition) GetNamespaces() *meta.LabelSelector {
	return this.namespaces
}
func (this *_Definition) GetFailurePolicy() adminreg.FailurePolicyType {
	if this.policy == "" {
		return adminreg.Fail
	}
	return this.policy
}
func (this *_Definition) GetOperations() []adminreg.OperationType {
	result := this.operations[:0:0]
	copy(result, this.operations)
	return result
}

func (this *_Definition) String() string {
	s := ""
	s += fmt.Sprintf("  namespaces: %+v\n", this.namespaces)
	s += fmt.Sprintf("  operations: %+v\n", this.operations)
	s += fmt.Sprintf("  failurePolicy: %+v\n", this.policy)
	return s
}

////////////////////////////////////////////////////////////////////////////////
// configuration
////////////////////////////////////////////////////////////////////////////////

type configuration struct {
	settings _Definition
}

var _ webhook.HandlerFactory = (*configuration)(nil)

func (this configuration) IgnoreFailures() configuration {
	this.settings.policy = adminreg.Ignore
	return this
}

func (this configuration) Operation(op ...adminreg.OperationType) configuration {
	this.settings.operations = append(this.settings.operations, op...)
	return this
}

func (this configuration) Namespaces(selector *meta.LabelSelector) configuration {
	this.settings.namespaces = selector
	return this
}

func (this configuration) CreateHandler() webhook.WebhookHandler {
	return &this.settings
}
