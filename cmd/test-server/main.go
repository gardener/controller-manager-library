// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/gardener/controller-manager-library/pkg/controllermanager"

	_ "github.com/gardener/controller-manager-library/pkg/controllermanager/examples/module/test"
	//	_ "github.com/gardener/gardener-botanist-aws/pkg/controller/controlplane"
	_ "github.com/gardener/controller-manager-library/pkg/controllermanager/examples/server/test"

	_ "github.com/gardener/controller-manager-library/pkg/resources/defaultscheme/v1.18"
)

func main() {
	controllermanager.Start("test-server", "Launch the Test Server", "A test server using the controller-manager-library")
}
