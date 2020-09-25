/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
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

	Validate(p webhook.Interface) error
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
	return &configuration{_Definition{htype, nil}}
}
