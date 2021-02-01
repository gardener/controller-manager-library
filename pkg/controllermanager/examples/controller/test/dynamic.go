/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package test

import (
	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller/reconcile"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
)

type dynamic struct {
	reconcile.DefaultReconciler
	controller controller.Interface
	notified   bool
}

var _ reconcile.Interface = &dynamic{}

func Dynamic(controller controller.Interface) (reconcile.Interface, error) {
	return &dynamic{
		controller: controller,
		notified:   true,
	}, nil
}

func (h *dynamic) Reconcile(logger logger.LogContext, obj resources.Object) reconcile.Status {

	if h.notified {
		//logger.Infof("GOT dynamic %s", obj.ClusterKey())
	} else {
		h.notified = true
		logger.Infof("GOT dynamic %s: %+#v\n", obj.GroupVersionKind(), obj.Data())
	}

	return reconcile.Succeeded(logger)
}
