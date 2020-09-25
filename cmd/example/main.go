/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"github.com/gardener/controller-manager-library/pkg/controllermanager"

	_ "github.com/gardener/controller-manager-library/pkg/controllermanager/examples/controller/example"
	//	_ "github.com/gardener/gardener-botanist-aws/pkg/controller/controlplane"
	_ "github.com/gardener/controller-manager-library/pkg/controllermanager/examples/webhook/conversion"

	_ "github.com/gardener/controller-manager-library/pkg/resources/defaultscheme/v1.12"
)

func main() {
	controllermanager.Start("example", "Launch the Exampler", "A conversion webhook and controller using the controller-manager-library")
}
