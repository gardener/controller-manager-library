/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"github.com/gardener/controller-manager-library/pkg/controllermanager"
)

func main() {
	controllermanager.Start("gardener-extension-controller-manager", "Launch the Gardener extension controller manager", "nothing")
}
