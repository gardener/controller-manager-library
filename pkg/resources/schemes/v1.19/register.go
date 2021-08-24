/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package v1_19 // golint: ignore

import (
	admissionregistration "k8s.io/api/admissionregistration/v1"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1beta1"
	networking "k8s.io/api/networking/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/gardener/controller-manager-library/pkg/resources/schemes"
)

var (
	SchemeBuilder = runtime.NewSchemeBuilder(
		discovery.AddToScheme,
		core.AddToScheme,
		apps.AddToScheme,
		admissionregistration.AddToScheme,
		networking.AddToScheme,
		apiextensions.AddToScheme)
	AddToScheme  = SchemeBuilder.AddToScheme
	SchemeSource = schemes.SchemeFunctionSource(Scheme)

	scheme *runtime.Scheme
)

func init() {
	scheme = runtime.NewScheme()
	SchemeBuilder.AddToScheme(scheme)
}

func Scheme() *runtime.Scheme {
	return scheme
}
