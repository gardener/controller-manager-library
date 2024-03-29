/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package admission

import (
	"errors"
	"net/http"

	"gomodules.xyz/jsonpatch/v2"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/webhook"
	"github.com/gardener/controller-manager-library/pkg/logger"
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
	Handle(logger.LogContext, Request) Response
}

type AdmissionHandlerType func(wh webhook.Interface) (Interface, error)

// WebhookFunc implements Handler interface using a single function.
type WebhookFunc func(logger.LogContext, Request) Response

var _ Interface = WebhookFunc(nil)

// Handle process the AdmissionRequest by invoking the underlying function.
func (this WebhookFunc) Handle(logger logger.LogContext, req Request) Response {
	return this(logger, req)
}

func (this WebhookFunc) Type() AdmissionHandlerType {
	return func(webhook.Interface) (Interface, error) { return this, nil }
}

// DefaultHandler can be used for a default implementation of all interface
// methods
type DefaultHandler struct {
}

func (this *DefaultHandler) Handle(logger.LogContext, Request) Response {
	return Allowed("always")
}
