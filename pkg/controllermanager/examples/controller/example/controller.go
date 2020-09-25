/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package example

import (
	"time"

	"github.com/gardener/controller-manager-library/pkg/config"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/examples/apis/example/v1alpha1"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller/reconcile"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
)

func init() {
	controller.Configure("example").
		Reconciler(Create).RequireLease().
		DefaultWorkerPool(10, 0*time.Second).
		OptionsByExample("options", &Config{}).
		MainResource(v1alpha1.GroupName, "Example", controller.NamespaceSelection("default")).
		MustRegister()
}

type Config struct {
	message string
}

func (this *Config) AddOptionsToSet(set config.OptionSet) {
	set.AddStringOption(&this.message, "message", "", "handle", "message")
}

type reconciler struct {
	reconcile.DefaultReconciler
	controller controller.Interface
	config     *Config
}

var _ reconcile.Interface = &reconciler{}

///////////////////////////////////////////////////////////////////////////////

func Create(controller controller.Interface) (reconcile.Interface, error) {

	config, err := controller.GetOptionSource("options")
	if err == nil {
		controller.Infof("found message option: %s", config.(*Config).message)
	}

	return &reconciler{controller: controller, config: config.(*Config)}, nil
}

///////////////////////////////////////////////////////////////////////////////

func (h *reconciler) Reconcile(logger logger.LogContext, obj resources.Object) reconcile.Status {
	logger.Infof("%s example (%T)", h.config.message, obj.Data())
	return reconcile.Succeeded(logger)
}

func (h *reconciler) Delete(logger logger.LogContext, obj resources.Object) reconcile.Status {
	logger.Infof("should delete")
	return reconcile.Succeeded(logger)
}

func (h *reconciler) Deleted(logger logger.LogContext, key resources.ClusterObjectKey) reconcile.Status {
	logger.Infof("is deleted")
	return reconcile.Succeeded(logger)
}
