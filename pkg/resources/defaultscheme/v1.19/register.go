/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package v1_19 // golint: ignore

import (
	"github.com/gardener/controller-manager-library/pkg/utils"
	admissionregistration "k8s.io/api/admissionregistration/v1"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	. "github.com/gardener/controller-manager-library/pkg/resources/abstract"
)

func init() {
	utils.Must(Register(core.SchemeBuilder))
	utils.Must(Register(apps.SchemeBuilder))
	utils.Must(Register(admissionregistration.SchemeBuilder))
	utils.Must(Register(apiextensions.SchemeBuilder))
	utils.Must(Register(networking.SchemeBuilder))
}
