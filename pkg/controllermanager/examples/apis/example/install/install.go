/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package install

import (
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/examples/apis/example"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/examples/apis/example/v1alpha1"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/examples/apis/example/v1beta1"
	"github.com/gardener/controller-manager-library/pkg/resources/apiextensions"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/examples/apis/example/crds"
)

func init() {
	crds.AddToRegistry(apiextensions.DefaultRegistry())
}

// Install installs all APIs in the scheme.
func Install(scheme *runtime.Scheme) {
	utilruntime.Must(v1beta1.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	utilruntime.Must(example.AddToScheme(scheme))
}
