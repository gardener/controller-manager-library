/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"time"

	"github.com/gardener/controller-manager-library/pkg/config"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller/reconcile/reconcilers"
	"github.com/gardener/controller-manager-library/pkg/resources"
)

var secretGK = resources.NewGroupKind("core", "Secret")

func init() {
	controller.Configure("cm").
		Reconciler(Create).RequireLease().
		DefaultWorkerPool(2, 0*time.Second).
		Commands("poll").
		StringOption("test", "Controller argument").
		OptionSource("test", controller.OptionSourceCreator(&Config{})).
		MainResource("core", "ConfigMap", controller.NamespaceSelection("default")).
		With(reconcilers.SecretUsageReconciler(controller.CLUSTER_MAIN)).
		WorkerPool("dynamic", 2, 0).
		Reconciler(Dynamic, "dynamic").
		MustRegister()
}

type Config struct {
	option string
}

func (this *Config) AddOptionsToSet(set config.OptionSet) {
	set.AddStringOption(&this.option, "option", "", "", "2nd controller argument")
}
