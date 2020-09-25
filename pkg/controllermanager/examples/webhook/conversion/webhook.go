/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/examples/apis/example"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/examples/apis/example/install"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/webhook"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/webhook/conversion"
)

func init() {
	scheme := runtime.NewScheme()
	install.Install(scheme)

	webhook.Configure("example.gardener.cloud").
		Kind(conversion.SchemeBasedConversion()).
		Scheme(scheme).
		Cluster(webhook.CLUSTER_MAIN).
		Resource(example.GroupName, "Example").
		MustRegister()
}
