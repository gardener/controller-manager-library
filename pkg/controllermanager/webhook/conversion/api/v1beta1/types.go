/*
SPDX-FileCopyrightText: 2017 The Kubernetes Authors.

SPDX-License-Identifier: Apache-2.0
*/

package v1beta1

import (
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ConversionReview v1beta1.ConversionReview

/*
// ConversionReview describes a conversion request/response.
type ConversionReview struct {
	metav1.TypeMeta `json:",inline"`
	// request describes the attributes for the conversion request.
	// +optional
	Request *ConversionRequest `json:"request,omitempty" protobuf:"bytes,1,opt,name=request"`
	// response describes the attributes for the conversion response.
	// +optional
	Response *ConversionResponse `json:"response,omitempty" protobuf:"bytes,2,opt,name=response"`
}

// ConversionRequest describes the conversion request parameters.
type ConversionRequest struct {
	// uid is an identifier for the individual request/response. It allows distinguishing instances of requests which are
	// otherwise identical (parallel requests, etc).
	// The UID is meant to track the round trip (request/response) between the Kubernetes API server and the webhook, not the user request.
	// It is suitable for correlating log entries between the webhook and apiserver, for either auditing or debugging.
	UID types.UID `json:"uid" protobuf:"bytes,1,name=uid"`
	// desiredAPIVersion is the version to convert given objects to. e.g. "myapi.example.com/v1"
	DesiredAPIVersion string `json:"desiredAPIVersion" protobuf:"bytes,2,name=desiredAPIVersion"`
	// objects is the list of custom resource objects to be converted.
	Objects []runtime.RawExtension `json:"objects" protobuf:"bytes,3,rep,name=objects"`
}

// ConversionResponse describes a conversion response.
type ConversionResponse struct {
	// uid is an identifier for the individual request/response.
	// This should be copied over from the corresponding `request.uid`.
	UID types.UID `json:"uid" protobuf:"bytes,1,name=uid"`
	// convertedObjects is the list of converted version of `request.objects` if the `result` is successful, otherwise empty.
	// The webhook is expected to set `apiVersion` of these objects to the `request.desiredAPIVersion`. The list
	// must also have the same size as the input list with the same objects in the same order (equal kind, metadata.uid, metadata.name and metadata.namespace).
	// The webhook is allowed to mutate labels and annotations. Any other change to the metadata is silently ignored.
	ConvertedObjects []runtime.RawExtension `json:"convertedObjects" protobuf:"bytes,2,rep,name=convertedObjects"`
	// result contains the result of conversion with extra details if the conversion failed. `result.status` determines if
	// the conversion failed or succeeded. The `result.status` field is required and represents the success or failure of the
	// conversion. A successful conversion must set `result.status` to `Success`. A failed conversion must set
	// `result.status` to `Failure` and provide more details in `result.message` and return http status 200. The `result.message`
	// will be used to construct an error message for the end user.
	Result metav1.Status `json:"result" protobuf:"bytes,3,name=result"`
}
*/
