/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"github.com/gardener/controller-manager-library/pkg/controllermanager"

	//	_ "github.com/gardener/gardener-botanist-aws/pkg/controller/controlplane"
	_ "github.com/gardener/controller-manager-library/pkg/controllermanager/examples/webhook/test"

	_ "github.com/gardener/controller-manager-library/pkg/resources/defaultscheme/v1.18"
)

func main() {
	controllermanager.Start("test-webhook", "Launch the Test Controller", "A test webhook using the controller-manager-library")
}
