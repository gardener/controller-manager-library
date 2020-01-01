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
	"context"
	"encoding/json"
	"fmt"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"io"
	"io/ioutil"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"net/http"
)

var admissionScheme = runtime.NewScheme()
var admissionCodecs = serializer.NewCodecFactory(admissionScheme)
var defaultDecoder = NewDecoder(resources.DefaultScheme())

func init() {
	utilruntime.Must(admissionv1beta1.AddToScheme(admissionScheme))
}

var _ http.Handler = &HTTPHandler{}

// HTTPHandler represents each individual webhook.
type HTTPHandler struct {
	// Handler actually processes an admission request returning whether it was allowed or denied,
	// and potentially patches to apply to the handler.
	webhook Interface

	// decoder is constructed on receiving a scheme and passed down to then handler
	decoder *Decoder

	logger.LogContext
}

func New(logger logger.LogContext, scheme *runtime.Scheme, webhook Interface) *HTTPHandler {
	d := defaultDecoder
	if scheme != nil {
		d = NewDecoder(scheme)
	}
	return &HTTPHandler{webhook: webhook, decoder: d, LogContext: logger}
}

func (this *HTTPHandler) Webhook() Interface {
	return this.webhook
}

// handle processes AdmissionRequest.
// If the webhook is mutating type, it delegates the AdmissionRequest to each handler and merge the patches.
// If the webhook is validating type, it delegates the AdmissionRequest to each handler and
// deny the request if anyone denies.
func (this *HTTPHandler) handle(ctx context.Context, req Request) Response {
	this.Infof("handle request for %q (%s)", req.Name, req.Resource)
	resp := this.webhook.Handle(ctx, req)
	if err := resp.Complete(req); err != nil {
		this.Error(err, "unable to encode response")
		return ErrorResponse(http.StatusInternalServerError, errUnableToEncodeResponse)
	}
	return resp
}

// GetDecoder returns a decoder to decode the objects embedded in admission requests.
// It may be nil if we haven't received a scheme to use to determine object types yet.
func (this *HTTPHandler) GetDecoder() *Decoder {
	return this.decoder
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
	reviewResponse = this.handle(context.Background(), req)
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
