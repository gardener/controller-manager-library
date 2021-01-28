/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package test

import (
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller/reconcile"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller/reconcile/reconcilers"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
)

type reconciler struct {
	reconcile.DefaultReconciler
	controller controller.Interface
	secrets    *reconcilers.SecretUsageCache
}

var _ reconcile.Interface = &reconciler{}

func Create(controller controller.Interface) (reconcile.Interface, error) {
	val, err := controller.GetStringOption("test")
	if err == nil {
		controller.Infof("found option test: %s", val)
	}

	config, err := controller.GetOptionSource("test")
	if err == nil {
		controller.Infof("found option option: %s", config.(*Config).option)
	}

	return &reconciler{
		controller: controller,
		secrets:    reconcilers.AccessSecretUsageCache(controller),
	}, nil
}

///////////////////////////////////////////////////////////////////////////////

func (h *reconciler) Setup() error {
	h.controller.Infof("setup of reconciler")
	return nil
}

func (h *reconciler) Start() {
	h.controller.EnqueueCommand("poll")
	h.controller.Infof("registering dynamic resources")
	w := controller.NewWatch(controller.CLUSTER_MAIN, "dynamic", "dynamic", controller.NewResourceKey("core", "Pod"))
	h.controller.RegisterWatch(w)
}

func (h *reconciler) Command(logger logger.LogContext, cmd string) reconcile.Status {
	logger.Infof("got command %q", cmd)
	return reconcile.Succeeded(logger).RescheduleAfter(10 * time.Second)
}

func (h *reconciler) Reconcile(logger logger.LogContext, obj resources.Object) reconcile.Status {
	/*
		o, err := resources.UnstructuredObject(obj)
		if err != nil {
			return reconcile.Failed(logger, err)
		}
		logger.Infof("GOT %s: %+#v\n", o.GroupVersionKind(), o.Data())
	*/
	switch o := obj.Data().(type) {
	case *corev1.ConfigMap:
		return h.reconcileConfigMap(logger, obj.ClusterKey(), o)
	}

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
	h.secrets.SetUsesFor(key, nil)
	return reconcile.Succeeded(logger)
}

func (h *reconciler) secretRef(namespace, name string) *resources.ClusterObjectKey {
	if strings.HasPrefix(name, "secret") {
		name = name[6:]
		if name != "" {
			key := resources.NewClusterKey(h.controller.GetMainCluster().GetId(), secretGK, namespace, name)
			return &key
		}
	}
	return nil
}

func (h *reconciler) reconcileConfigMap(logger logger.LogContext, key resources.ClusterObjectKey, configMap *corev1.ConfigMap) reconcile.Status {
	logger.Infof("should reconcile configmap")
	// Example how to add to workqueue
	// resources, _ := h.controller.Data(configMap)
	// key, _ := controller.ObjectKeyFunc(resources)
	//	h.controller.GetWorkqueue().Add(key)

	secret := h.secretRef(configMap.Namespace, configMap.Name)
	h.secrets.SetUsesFor(key, secret)

	return reconcile.Succeeded(logger)
}
