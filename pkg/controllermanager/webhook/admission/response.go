/*
SPDX-FileCopyrightText: 2018 The Kubernetes Authors.

SPDX-License-Identifier: Apache-2.0
*/

package admission

import (
	"net/http"

	"gomodules.xyz/jsonpatch/v2"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Allowed constructs a response indicating that the given operation
// is allowed (without any patches).
func Allowed(reason string) Response {
	return ValidationResponse(true, reason)
}

// Denied constructs a response indicating that the given operation
// is not allowed.
func Denied(reason string) Response {
	return ValidationResponse(false, reason)
}

// Patched constructs a response indicating that the given operation is
// allowed, and that the target object should be modified by the given
// JSONPatch operations.
func Patched(reason string, patches ...jsonpatch.JsonPatchOperation) Response {
	resp := Allowed(reason)
	resp.Patches = patches

	return resp
}

// ErrorResponse creates a new Response for error-handling a request.
func ErrorResponse(code int32, err error) Response {
	return Response{
		AdmissionResponse: admissionv1beta1.AdmissionResponse{
			Allowed: false,
			Result: &meta.Status{
				Code:    code,
				Message: err.Error(),
			},
		},
	}
}

// ValidationResponse returns a response for admitting a request.
func ValidationResponse(allowed bool, reason string) Response {
	code := http.StatusForbidden
	if allowed {
		code = http.StatusOK
	}
	resp := Response{
		AdmissionResponse: admissionv1beta1.AdmissionResponse{
			Allowed: allowed,
			Result: &meta.Status{
				Code: int32(code),
			},
		},
	}
	if len(reason) > 0 {
		resp.Result.Reason = meta.StatusReason(reason)
	}
	return resp
}

// PatchResponseFromRaw takes 2 byte arrays and returns a new response with json patch.
// The original object should be passed in as raw bytes to avoid the roundtripping problem
// described in https://github.com/kubernetes-sigs/kubebuilder/issues/510.
func PatchResponseFromRaw(original, current []byte) Response {
	patches, err := jsonpatch.CreatePatch(original, current)
	if err != nil {
		return ErrorResponse(http.StatusInternalServerError, err)
	}
	return Response{
		Patches: patches,
		AdmissionResponse: admissionv1beta1.AdmissionResponse{
			Allowed:   true,
			PatchType: func() *admissionv1beta1.PatchType { pt := admissionv1beta1.PatchTypeJSONPatch; return &pt }(),
		},
	}
}
