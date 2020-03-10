/*
 * Copyright 2020 SAP SE or an SAP affiliate company. All rights reserved.
 * This file is licensed under the Apache Software License, v. 2 except as noted
 * otherwise in the LICENSE file
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 *
 */

package test

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

func (h *reconciler) Start() {
	h.controller.EnqueueCommand("poll")
}

func (h *reconciler) Command(logger logger.LogContext, cmd string) reconcile.Status {
	logger.Infof("got command %q", cmd)
	return reconcile.Succeeded(logger).RescheduleAfter(10 * time.Second)
}

func (h *reconciler) Reconcile(logger logger.LogContext, obj resources.Object) reconcile.Status {
	logger.Infof("%s example (%T)", h.config.message, obj.Data())
	return reconcile.Succeeded(logger)
}

func (h *reconciler) Delete(logger logger.LogContext, obj resources.Object) reconcile.Status {
	// logger.Infof("delete infrastructure %s", resources.Description(obj))
	logger.Infof("should delete")
	return reconcile.Succeeded(logger)
}

func (h *reconciler) Deleted(logger logger.LogContext, key resources.ClusterObjectKey) reconcile.Status {
	// logger.Infof("delete infrastructure %s", resources.Description(obj))
	logger.Infof("is deleted")
	return reconcile.Succeeded(logger)
}
