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

package conversion

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/webhook/conversion/api"
	"github.com/gardener/controller-manager-library/pkg/logger"
)

var _ http.Handler = &HTTPHandler{}

// HTTPHandler represents each individual webhook.
type HTTPHandler struct {
	// Handler actually processes a conversion review request,
	webhook Interface

	logger.LogContext
}

func (this *HTTPHandler) Webhook() Interface {
	return this.webhook
}

// handle processes ConversionReviewRequest.
func (this *HTTPHandler) handle(req *Request) *Response {
	logctx := this.NewContext("conversion", req.DesiredAPIVersion)
	logctx.Infof("handle request for %d resources", len(req.Objects))
	resp := &Response{
		UID:              req.UID,
		ConvertedObjects: make([]runtime.RawExtension, len(req.Objects)),
		Result: meta.Status{
			Status:  SUCCESS,
			Message: "",
			Reason:  "",
			Details: nil,
			Code:    http.StatusOK,
		},
	}
	for i, o := range req.Objects {
		c, err := this.webhook.Handle(logctx, req.DesiredAPIVersion, o)
		if err != nil {
			return ErrorResponse(req, http.StatusBadRequest, err)
		}
		r, err := json.Marshal(c)
		if err != nil {
			return ErrorResponse(req, http.StatusUnprocessableEntity, err)
		}
		resp.ConvertedObjects[i] = runtime.RawExtension{
			Raw: r,
		}
	}
	return resp
}

func (this *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	obj, response := this.serveHTTP(r)

	if obj == nil {
		w.WriteHeader(int(response.Result.Code))
		return
	}
	answer := &api.ConversionReview{
		Response: response,
	}
	answer.SetGroupVersionKind(obj.GetObjectKind().GroupVersionKind())

	err := reviewScheme.Convert(answer, obj, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(obj)
	if err != nil {
		this.Errorf("failed to write response: %s", err)
		return
	}
}

func (this *HTTPHandler) serveHTTP(r *http.Request) (runtime.Object, *Response) {
	var body []byte
	var err error

	if r.Body != nil {
		if body, err = ioutil.ReadAll(r.Body); err != nil {
			this.Error(err, "unable to read the body from the incoming request")
			return nil, ErrorResponse(nil, http.StatusBadRequest, err)
		}
	} else {
		err = fmt.Errorf("request body is empty")
		this.Error(err)
		return nil, ErrorResponse(nil, http.StatusBadRequest, err)
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		err = fmt.Errorf("contentType=%s, expected application/json", contentType)
		this.Errorf("unable to process a request with an unknown content type: %s", contentType)
		return nil, ErrorResponse(nil, http.StatusBadRequest, err)
	}

	versions := &runtime.VersionedObjects{}

	if err := reviewDecoder.DecodeInto(body, versions); err != nil {
		this.Errorf("unable to decode the request", err)
		return nil, ErrorResponse(nil, http.StatusBadRequest, err)
	}

	// TODO: add panic-recovery for Handle
	return versions.First(), this.handle(versions.Last().(*api.ConversionReview).Request)
}
