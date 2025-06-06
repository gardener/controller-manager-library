//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors

SPDX-License-Identifier: Apache-2.0
*/

// Code generated by conversion-gen. DO NOT EDIT.

package v1beta1

import (
	unsafe "unsafe"

	api "github.com/gardener/controller-manager-library/pkg/controllermanager/webhook/conversion/api"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	conversion "k8s.io/apimachinery/pkg/conversion"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

func init() {
	localSchemeBuilder.Register(RegisterConversions)
}

// RegisterConversions adds conversion functions to the given scheme.
// Public to allow building arbitrary schemes.
func RegisterConversions(s *runtime.Scheme) error {
	if err := s.AddGeneratedConversionFunc((*ConversionReview)(nil), (*api.ConversionReview)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1beta1_ConversionReview_To_api_ConversionReview(a.(*ConversionReview), b.(*api.ConversionReview), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*api.ConversionReview)(nil), (*ConversionReview)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_api_ConversionReview_To_v1beta1_ConversionReview(a.(*api.ConversionReview), b.(*ConversionReview), scope)
	}); err != nil {
		return err
	}
	return nil
}

func autoConvert_v1beta1_ConversionReview_To_api_ConversionReview(in *ConversionReview, out *api.ConversionReview, s conversion.Scope) error {
	out.Request = (*api.ConversionRequest)(unsafe.Pointer(in.Request))
	out.Response = (*api.ConversionResponse)(unsafe.Pointer(in.Response))
	return nil
}

// Convert_v1beta1_ConversionReview_To_api_ConversionReview is an autogenerated conversion function.
func Convert_v1beta1_ConversionReview_To_api_ConversionReview(in *ConversionReview, out *api.ConversionReview, s conversion.Scope) error {
	return autoConvert_v1beta1_ConversionReview_To_api_ConversionReview(in, out, s)
}

func autoConvert_api_ConversionReview_To_v1beta1_ConversionReview(in *api.ConversionReview, out *ConversionReview, s conversion.Scope) error {
	out.Request = (*apiextensionsv1beta1.ConversionRequest)(unsafe.Pointer(in.Request))
	out.Response = (*apiextensionsv1beta1.ConversionResponse)(unsafe.Pointer(in.Response))
	return nil
}

// Convert_api_ConversionReview_To_v1beta1_ConversionReview is an autogenerated conversion function.
func Convert_api_ConversionReview_To_v1beta1_ConversionReview(in *api.ConversionReview, out *ConversionReview, s conversion.Scope) error {
	return autoConvert_api_ConversionReview_To_v1beta1_ConversionReview(in, out, s)
}
