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
	"net/http"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/webhook"
)

type Definition interface {
	GetKind() webhook.WebhookKind
	GetHTTPHandler(wh webhook.Interface) (http.Handler, error)

	Validation(p webhook.Interface) error
}

type _Definition struct {
	factory   ConversionHandlerType
	validator webhook.WebhookValidator
}

var _ webhook.WebhookHandler = (*_Definition)(nil)
var _ Definition = (*_Definition)(nil)

func (this *_Definition) GetKind() webhook.WebhookKind {
	return webhook.CONVERTING
}

func (this *_Definition) GetHTTPHandler(wh webhook.Interface) (http.Handler, error) {
	h, err := this.factory(wh)
	if err != nil {
		return nil, err
	}
	return &HTTPHandler{webhook: h, LogContext: wh}, nil
}

func (this *_Definition) String() string {
	return ""
}

func (this *_Definition) Validate(wh webhook.Interface) error {
	if this.validator == nil {
		return nil
	}
	return this.validator.Validate(wh)
}

////////////////////////////////////////////////////////////////////////////////
// configuration
////////////////////////////////////////////////////////////////////////////////

type configuration struct {
	settings _Definition
}

var _ webhook.HandlerFactory = (*configuration)(nil)

func (this configuration) CreateHandler() webhook.WebhookHandler {
	return &this.settings
}

func Conversion(htype ConversionHandlerType) *configuration {
	return &configuration{_Definition{htype}}
}
