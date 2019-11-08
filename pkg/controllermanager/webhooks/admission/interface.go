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
	"net/http"
)

import (
	"context"
	"errors"
	"github.com/appscode/jsonpatch"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
)

var (
	errUnableToEncodeResponse = errors.New("unable to encode response")
)

// Request defines the input for an admission handler.
// It contains information to identify the object in
// question (group, version, kind, resource, subresource,
// name, namespace), as well as the operation in question
// (e.g. Get, Create, etc), and the object itself.
type Request struct {
	admissionv1beta1.AdmissionRequest
}

// Response is the output of an admission handler.
// It contains a response indicating if a given
// operation is allowed, as well as a set of patches
// to mutate the object in the case of a mutating admission handler.
type Response struct {
	// Patches are the JSON patches for mutating webhooks.
	// Using this instead of setting Response.Patch to minimize
	// overhead of serialization and deserialization.
	// Patches set here will override any patches in the response,
	// so leave this empty if you want to set the patch response directly.
	Patches []jsonpatch.JsonPatchOperation
	// AdmissionResponse is the raw admission response.
	// The Patch field in it will be overwritten by the listed patches.
	admissionv1beta1.AdmissionResponse
}

// Complete populates any fields that are yet to be set in
// the underlying AdmissionResponse, It mutates the response.
func (this *Response) Complete(req Request) error {
	this.UID = req.UID

	// ensure that we have a valid status code
	if this.Result == nil {
		this.Result = &metav1.Status{}
	}
	if this.Result.Code == 0 {
		this.Result.Code = http.StatusOK
	}

	if len(this.Patches) == 0 {
		return nil
	}

	var err error
	this.Patch, err = json.Marshal(this.Patches)
	if err != nil {
		return err
	}
	patchType := admissionv1beta1.PatchTypeJSONPatch
	this.PatchType = &patchType

	return nil
}

// Interface can handle an AdmissionRequest.
type Interface interface {
	Handle(context.Context, Request) Response
}

// WebhookFunc implements Handler interface using a single function.
type WebhookFunc func(context.Context, Request) Response

var _ Interface = WebhookFunc(nil)

// Handle process the AdmissionRequest by invoking the underlying function.
func (this WebhookFunc) Handle(ctx context.Context, req Request) Response {
	return this(ctx, req)
}
