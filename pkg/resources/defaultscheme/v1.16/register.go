/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package v1_16 // golint: ignore

import (
	. "github.com/gardener/controller-manager-library/pkg/resources/abstract"
	v1_16 "github.com/gardener/controller-manager-library/pkg/resources/schemes/v1.16"
)

func init() {
	Register(v1_16.SchemeBuilder)
}
