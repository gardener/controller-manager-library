/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package conversion

import (
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/webhook"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/webhook/conversion/api"
	"github.com/gardener/controller-manager-library/pkg/logger"
)

type Request = api.ConversionRequest

type Response = api.ConversionResponse

type Interface interface {
	Handle(log logger.LogContext, version string, obj runtime.RawExtension) (runtime.Object, error)
}

type ConversionHandlerType func(wh webhook.Interface) (Interface, error)

// WebhookFunc implements Handler interface using a single function.
type WebhookFunc func(log logger.LogContext, version string, obj runtime.RawExtension) (runtime.Object, error)

var _ Interface = WebhookFunc(nil)

// Handle process the AdmissionRequest by invoking the underlying function.
func (this WebhookFunc) Handle(log logger.LogContext, version string, obj runtime.RawExtension) (runtime.Object, error) {
	return this(log, version, obj)
}

func (this WebhookFunc) Type() ConversionHandlerType {
	return func(webhook.Interface) (Interface, error) { return this, nil }
}
