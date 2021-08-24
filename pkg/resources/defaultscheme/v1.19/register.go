/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package v1_19 // golint: ignore

import (
	. "github.com/gardener/controller-manager-library/pkg/resources/abstract"
	v1_19 "github.com/gardener/controller-manager-library/pkg/resources/schemes/v1.19"
)

func init() {
	Register(v1_19.SchemeBuilder)
}
