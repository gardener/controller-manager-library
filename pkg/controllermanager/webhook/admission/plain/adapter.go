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
	"fmt"
	"net/http"

	"github.com/gardener/controller-manager-library/pkg/logger"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/webhook"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/webhook/admission"
	resources "github.com/gardener/controller-manager-library/pkg/resources/plain"
)

type AdmissionHandlerType func(webhook.Interface) (Interface, error)

type Adapter struct {
	resources resources.Resources
	webhook   webhook.Interface
	handler   Interface
}

var _ admission.Interface = &Adapter{}

func Adapt(c AdmissionHandlerType) admission.AdmissionHandlerType {

	return func(webhook webhook.Interface) (admission.Interface, error) {
		r := resources.NewResourceContext(webhook.GetContext(), webhook.GetScheme()).Resources()
		h, err := c(webhook)
		if err != nil {
			return nil, err
		}
		return &Adapter{
			resources: r,
			webhook:   webhook,
			handler:   h,
		}, nil
	}
}

func (this *Adapter) Handle(logger logger.LogContext, request admission.Request) admission.Response {
	var object resources.Object
	var old resources.Object
	var err error

	if request.Object.Raw != nil {
		object, err = this.resources.Decode(request.Object.Raw)
		if err != nil {
			return admission.ErrorResponse(http.StatusInternalServerError, fmt.Errorf("cannot parse object", err))
		}
		request.Object.Object = object.Data()
	}
	if request.OldObject.Raw != nil {
		old, err = this.resources.Decode(request.OldObject.Raw)
		if err != nil {
			return admission.ErrorResponse(http.StatusInternalServerError, fmt.Errorf("cannot parse oldobject", err))
		}
		request.OldObject.Object = old.Data()
	}
	mapped := Request{
		Request:   request,
		Object:    object,
		OldObject: old,
	}
	return this.handler.Handle(logger, mapped)
}
