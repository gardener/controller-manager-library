/*
 * SPDX-FileCopyrightText: Copyright Contributors to the Gardener project
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
