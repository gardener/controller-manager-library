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
