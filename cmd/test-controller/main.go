package main

import (
	"github.com/gardener/controller-manager-library/pkg/controllermanager"

	//	_ "github.com/gardener/gardener-botanist-aws/pkg/controller/controlplane"
	_ "github.com/gardener/controller-manager-library/pkg/controllermanager/test"
)

func main() {
	controllermanager.Start("test-controller", "Launch the Test Controller", "A test controller using the controller-manager-library")
}
