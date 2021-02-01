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
	"github.com/gardener/controller-manager-library/pkg/goutils"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
)

type reconciler struct {
	reconcile.DefaultReconciler
	controller controller.Interface
	secrets    *reconcilers.SecretUsageCache
	config     *Config
	dynamic    controller.Watch
	count      int
	last       goutils.GoRoutines
	lastadded  goutils.GoRoutines
}

var _ reconcile.Interface = &reconciler{}

func Create(cntr controller.Interface) (reconcile.Interface, error) {
	val, err := cntr.GetStringOption("test")
	if err == nil {
		cntr.Infof("found option test: %s", val)
	}

	config, err := cntr.GetOptionSource("test")
	if err == nil {
		cntr.Infof("found interval option: %d", config.(*Config).interval)
	}

	return &reconciler{
		controller: cntr,
		config:     config.(*Config),
		secrets:    reconcilers.AccessSecretUsageCache(cntr),
		dynamic:    controller.NewWatch(controller.CLUSTER_MAIN, "dynamic", "dynamic", controller.NewResourceKey("core", "Pod")),
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
}

func (h *reconciler) Command(logger logger.LogContext, cmd string) reconcile.Status {

	h.count++
	logger.Infof("got command %q (%d): %d goroutines", cmd, h.count, goutils.NumberOfGoRoutines())
	cur := goutils.ListGoRoutines(true)
	if h.last != nil {
		add, del := goutils.GoRoutineDiff(h.last, cur)
		for _, r := range add {
			logger.Infof("added   %3d [%s] %s", r.Id, r.Status, r.Current.Location)
			logger.Infof("            %s", r.First.Location)
		}
		for _, r := range del {
			logger.Infof("deleted %3d [%s] %s", r.Id, r.Status, r.Current.Location)
			logger.Infof("            %s", r.First.Location)
		}
		if len(del) > 0 && h.lastadded != nil {
			add, del := goutils.GoRoutineDiff(h.lastadded, del)
			h.lastadded = nil
			if len(add)+len(del) > 0 {
				logger.Warnf("*********************  mismatch in add/del of goroutines *******************")
				for _, r := range add {
					logger.Infof("additional  %3d [%s]", r.Id, r.Status)
					logger.Infof("      creator  %s", r.Creator.Location)
					for _, s := range r.Stack {
						logger.Infof("      stack    %s", s.Location)
					}
				}
				for _, r := range del {
					logger.Infof("missing     %3d [%s]", r.Id, r.Status)
					logger.Infof("      creator  %s", r.Creator.Location)
					for _, s := range r.Stack {
						logger.Infof("      stack    %s", s.Location)
					}
				}
			}
		}
		if len(add) > 0 {
			h.lastadded = add
		}
	}
	h.last = cur

	switch h.count % (h.config.interval * 2) {
	case h.config.interval:
		h.controller.RegisterWatch(h.dynamic)
	case 0:
		h.controller.UnregisterWatch(h.dynamic)
	}
	return reconcile.Succeeded(logger).RescheduleAfter(10 * time.Second)
}

func (h *reconciler) Reconcile(logger logger.LogContext, obj resources.Object) reconcile.Status {
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
	//logger.Infof("should reconcile configmap")
	// Example how to add to workqueue
	// resources, _ := h.controller.Data(configMap)
	// key, _ := controller.ObjectKeyFunc(resources)
	//	h.controller.GetWorkqueue().Add(key)

	secret := h.secretRef(configMap.Namespace, configMap.Name)
	h.secrets.SetUsesFor(key, secret)

	return reconcile.Succeeded(logger)
}
