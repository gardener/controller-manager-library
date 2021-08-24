// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/gardener/controller-manager-library/pkg/controllermanager"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"github.com/gardener/controller-manager-library/pkg/resources/schemes"
	v1_16 "github.com/gardener/controller-manager-library/pkg/resources/schemes/v1.16"
	v1_18 "github.com/gardener/controller-manager-library/pkg/resources/schemes/v1.18"

	//	_ "github.com/gardener/gardener-botanist-aws/pkg/controller/controlplane"
	_ "github.com/gardener/controller-manager-library/pkg/controllermanager/examples/controller/test"
)

func init() {
	resources.SetDefaultSchemeSource(
		schemes.NewFlavoredSchemeSource(schemes.Conditional(schemes.APIServerVersion("<v1.18"), v1_16.SchemeSource), v1_18.SchemeSource),
	)
}

func main() {
	controllermanager.Start("test-controller", "Launch the Test Controller", "A test controller using the controller-manager-library")
}
