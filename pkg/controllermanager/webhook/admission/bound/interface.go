/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package bound

import (
	"github.com/gardener/controller-manager-library/pkg/controllermanager/webhook"
	"github.com/gardener/controller-manager-library/pkg/logger"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/webhook/admission"
	"github.com/gardener/controller-manager-library/pkg/resources"
)

// Request describes the admission.Attributes for the admission request.
type Request struct {
	Request   admission.Request
	Object    resources.Object
	OldObject resources.Object
}

// Interface can handle an AdmissionRequest.
type Interface interface {
	Handle(logger.LogContext, Request) admission.Response
}

type AdmissionHandlerType func(webhook.Interface) (Interface, error)

// WebhookFunc implements Handler interface using a single function.
type WebhookFunc func(logger.LogContext, Request) admission.Response

var _ Interface = WebhookFunc(nil)

// Handle process the AdmissionRequest by invoking the underlying function.
func (this WebhookFunc) Handle(logger logger.LogContext, req Request) admission.Response {
	return this(logger, req)
}

func (this WebhookFunc) Type() AdmissionHandlerType {
	return func(webhook.Interface) (Interface, error) { return this, nil }
}

// DefaultHandler can be used for a default implementation of all interface
// methods
type DefaultHandler struct {
}

func (this *DefaultHandler) Handle(logger.LogContext, Request) admission.Response {
	return admission.Allowed("always")
}
