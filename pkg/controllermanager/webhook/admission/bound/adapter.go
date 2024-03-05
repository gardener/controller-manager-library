/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package bound

import (
	"fmt"
	"net/http"

	"github.com/gardener/controller-manager-library/pkg/logger"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/webhook"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/webhook/admission"
	"github.com/gardener/controller-manager-library/pkg/resources"
)

type Adapter struct {
	resources resources.Resources
	webhook   webhook.Interface
	handler   Interface
}

var _ admission.Interface = &Adapter{}

func Adapt(c AdmissionHandlerType) admission.AdmissionHandlerType {
	return func(webhook webhook.Interface) (admission.Interface, error) {
		cluster := webhook.GetCluster()
		if cluster == nil {
			return nil, fmt.Errorf("cluster required for regular resource abstraction")
		}
		r := cluster.Resources()
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
			return admission.ErrorResponse(http.StatusInternalServerError, fmt.Errorf("cannot parse object: %s", err))
		}
		request.Object.Object = object.Data()
	}
	if request.OldObject.Raw != nil {
		old, err = this.resources.Decode(request.OldObject.Raw)
		if err != nil {
			return admission.ErrorResponse(http.StatusInternalServerError, fmt.Errorf("cannot parse oldobject: %s", err))
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
