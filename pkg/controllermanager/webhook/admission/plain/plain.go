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

package plain

import (
	"github.com/gardener/controller-manager-library/pkg/logger"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/webhook/admission"
	"github.com/gardener/controller-manager-library/pkg/resources/plain"
)

// Request describes the admission.Attributes for the admission request.
type Request struct {
	Request   admission.Request
	Object    plain.Object
	OldObject plain.Object
}

// Interface can handle an AdmissionRequest.
type Interface interface {
	Handle(logger.LogContext, Request) admission.Response
}

// WebhookFunc implements Handler interface using a single function.
type WebhookFunc func(logger.LogContext, Request) admission.Response

var _ Interface = WebhookFunc(nil)

// Handle process the AdmissionRequest by invoking the underlying function.
func (this WebhookFunc) Handle(logger logger.LogContext, req Request) admission.Response {
	return this(logger, req)
}

// DefaultHandler can be used for a default implementation of all interface
// methods
type DefaultHandler struct {
}

func (this *DefaultHandler) Handle(logger.LogContext, Request) admission.Response {
	return admission.Allowed("always")
}
