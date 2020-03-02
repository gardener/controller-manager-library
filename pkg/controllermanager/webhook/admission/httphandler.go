/*
 * Copyright 2019 SAP SE or an SAP affiliate company. All rights reserved.
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

package admission

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"

	admissionv1beta1 "k8s.io/api/admission/v1beta1"
)

var _ http.Handler = &HTTPHandler{}

// HTTPHandler represents each individual webhook.
type HTTPHandler struct {
	// Handler actually processes an admission request returning whether it was allowed or denied,
	// and potentially patches to apply to the handler.
	webhook Interface

	logger.LogContext
}

func (this *HTTPHandler) Webhook() Interface {
	return this.webhook
}

// handle processes AdmissionRequest.
// If the webhook is mutating type, it delegates the AdmissionRequest to each handler and merge the patches.
// If the webhook is validating type, it delegates the AdmissionRequest to each handler and
// deny the request if anyone denies.
func (this *HTTPHandler) handle(req Request) Response {
	name := resources.NewObjectName(req.Namespace, req.Name)
	logctx := this.NewContext("object", name.String())
	logctx.Infof("handle request for %s", req.Resource)
	resp := this.webhook.Handle(logctx, req)
	if err := resp.Complete(req); err != nil {
		logctx.Error(err, "unable to encode response")
		return ErrorResponse(http.StatusInternalServerError, errUnableToEncodeResponse)
	}
	return resp
}

func (this *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var body []byte
	var err error

	var reviewResponse Response
	if r.Body != nil {
		if body, err = ioutil.ReadAll(r.Body); err != nil {
			this.Error(err, "unable to read the body from the incoming request")
			reviewResponse = ErrorResponse(http.StatusBadRequest, err)
			this.writeResponse(w, reviewResponse)
			return
		}
	} else {
		err = fmt.Errorf("request body is empty")
		this.Error(err)
		reviewResponse = ErrorResponse(http.StatusBadRequest, err)
		this.writeResponse(w, reviewResponse)
		return
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		err = fmt.Errorf("contentType=%s, expected application/json", contentType)
		this.Errorf("unable to process a request with an unknown content type: %s", contentType)
		reviewResponse = ErrorResponse(http.StatusBadRequest, err)
		this.writeResponse(w, reviewResponse)
		return
	}

	req := Request{}
	ar := admissionv1beta1.AdmissionReview{
		// avoid an extra copy
		Request: &req.AdmissionRequest,
	}
	if _, _, err := admissionCodecs.UniversalDeserializer().Decode(body, nil, &ar); err != nil {
		this.Errorf("unable to decode the request", err)
		reviewResponse = ErrorResponse(http.StatusBadRequest, err)
		this.writeResponse(w, reviewResponse)
		return
	}

	// TODO: add panic-recovery for Handle
	reviewResponse = this.handle(req)
	this.writeResponse(w, reviewResponse)
}

func (this *HTTPHandler) writeResponse(w io.Writer, response Response) {
	encoder := json.NewEncoder(w)
	responseAdmissionReview := admissionv1beta1.AdmissionReview{
		Response: &response.AdmissionResponse,
	}
	err := encoder.Encode(responseAdmissionReview)
	if err != nil {
		this.Errorf("unable to encode the response: %s", err)
		this.writeResponse(w, ErrorResponse(http.StatusInternalServerError, err))
	}
}
