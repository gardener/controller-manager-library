/*
 * Copyright 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package controllermanager

import (
	"context"
	"fmt"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"github.com/gardener/controller-manager-library/pkg/resources/access"
	"github.com/gardener/controller-manager-library/pkg/utils"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/config"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller"
	"github.com/gardener/controller-manager-library/pkg/ctxutil"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/server"
)

type ControllerManager struct {
	lock sync.Mutex
	controller.SharedAttributes

	name       string
	definition *Definition

	ctx           context.Context
	config        *config.Config
	clusters      cluster.Clusters
	registrations controller.Registrations
	plain_groups  map[string]StartupGroup
	lease_groups  map[string]StartupGroup
	controllers   Controllers
	prepared      map[string]*SyncPoint
	//shared_options map[string]*config.ArbitraryOption
}

var _ controller.Environment = &ControllerManager{}

type Controller interface {
	logger.LogContext

	GetName() string
	Owning() controller.ResourceKey
	GetDefinition() controller.Definition
	GetClusterHandler(name string) (*controller.ClusterHandler, error)

	Check() error
	Prepare() error
	Run()
}

func NewControllerManager(ctx context.Context, def *Definition) (*ControllerManager, error) {
	config := config.Get(ctx)
	ctx = context.WithValue(ctx, resources.ATTR_EVENTSOURCE, def.GetName())

	for n := range def.controller_defs.Names() {
		for _, r := range def.controller_defs.Get(n).RequiredControllers() {
			if def.controller_defs.Get(r) == nil {
				return nil, fmt.Errorf("controller %q requires controller %q, which is not declared", n, r)
			}
		}
	}

	if config.NamespaceRestriction && config.DisableNamespaceRestriction {
		log.Fatalf("contradiction options given for namespace restriction")
	}
	if !config.DisableNamespaceRestriction {
		config.NamespaceRestriction = true
	}
	config.DisableNamespaceRestriction = false

	if config.NamespaceRestriction {
		logger.Infof("enable namespace restriction for access control")
		access.RegisterNamespaceOnlyAccess()
	} else {
		logger.Infof("disable namespace restriction for access control")
	}
	if config.Namespace == "" {
		n := os.Getenv("NAMESPACE")
		if n != "" {
			config.Namespace = n
		} else {
			f := "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
			bytes, err := ioutil.ReadFile(f)
			if err == nil {
				n = string(bytes)
				n = strings.TrimSpace(n)
				if n != "" {
					config.Namespace = n

				}
			}
		}
	}

	name := def.GetName()
	if config.Name != "" {
		name = config.Name
	}

	if config.Namespace == "" {
		config.Namespace = "kube-system"
	}
	groups := def.Groups()

	logger.Infof("configured groups: %s", groups.AllGroups())

	if def.ControllerDefinitions().Size() == 0 {
		return nil, fmt.Errorf("no controller registered")
	}

	logger.Infof("configured controllers: %s", def.ControllerDefinitions().Names())

	active, err := groups.Activate(strings.Split(config.Controllers, ","))
	if err != nil {
		return nil, err
	}

	added := utils.StringSet{}
	for c := range active {
		req, err := def.controller_defs.GetRequiredControllers(c)
		if err != nil {
			return nil, err
		}
		added.AddSet(req)
	}
	added, _ = active.DiffFrom(added)
	if len(added) > 0 {
		logger.Infof("controllers implied by activated controllers: %s", added)
		active.AddSet(added)
	}

	registrations, err := def.Registrations(active.AsArray()...)
	if err != nil {
		return nil, err
	}
	if len(registrations) == 0 {
		return nil, fmt.Errorf("no controller activated")
	}

	set, err := def.ControllerDefinitions().DetermineRequestedClusters(def.ClusterDefinitions(), registrations.Names())
	if err != nil {
		return nil, err
	}

	lgr := logger.New()
	clusters, err := def.ClusterDefinitions().CreateClusters(ctx, lgr, config, set)
	if err != nil {
		return nil, err
	}

	cm := &ControllerManager{
		SharedAttributes: controller.SharedAttributes{
			LogContext: lgr,
		},
		clusters: clusters,

		name:          name,
		definition:    def,
		config:        config,
		registrations: registrations,
		prepared:      map[string]*SyncPoint{},

		plain_groups: map[string]StartupGroup{},
		lease_groups: map[string]StartupGroup{},
	}

	ctx = logger.Set(ctxutil.SyncContext(ctx), lgr)
	ctx = context.WithValue(ctx, cmkey, cm)
	cm.ctx = ctx
	return cm, nil
}

func (this *ControllerManager) GetName() string {
	return this.name
}

func (this *ControllerManager) GetContext() context.Context {
	return this.ctx
}

func (this *ControllerManager) GetConfig() *config.Config {
	return this.config
}

func (this *ControllerManager) GetCluster(name string) cluster.Interface {
	return this.clusters.GetCluster(name)
}

func (this *ControllerManager) GetClusters() cluster.Clusters {
	return this.clusters
}

func (this *ControllerManager) Run() error {
	var err error
	this.Infof("run %s\n", this.name)

	if this.config.ServerPortHTTP > 0 {
		server.Serve(this.ctx, "", this.config.ServerPortHTTP)
	}

	for _, def := range this.registrations {
		lines := strings.Split(def.String(), "\n")
		this.Infof("creating %s", lines[0])
		for _, l := range lines[1:] {
			this.Info(l)
		}
		cmp, err := this.definition.GetMappingsFor(def.GetName())
		if err != nil {
			return err
		}
		cntr, err := controller.NewController(this, def, cmp)
		if err != nil {
			return err
		}

		if def.RequireLease() {
			this.getLeaseStartupGroup(cntr.GetMainCluster()).Add(cntr)
		} else {
			this.getPlainStartupGroup(cntr.GetMainCluster()).Add(cntr)
		}
		this.controllers = append(this.controllers, cntr)
		this.prepared[cntr.GetName()] = &SyncPoint{}
	}

	this.controllers, err = this.controllers.getOrder(this)
	if err != nil {
		return err
	}

	for _, cntr := range this.controllers {
		err := this.checkController(cntr)
		if err != nil {
			return err
		}
	}

	err = this.startGroups(this.plain_groups, this.lease_groups)
	if err != nil {
		return err
	}

	<-this.ctx.Done()
	this.Info("waiting for controllers to shutdown")
	ctxutil.SyncPointWait(this.ctx, 120*time.Second)
	this.Info("exit controller manager")
	return nil
}

// checkController does all the checks that might cause startController to fail
// after the check startController can execute without error
func (this *ControllerManager) checkController(cntr Controller) error {
	return cntr.Check()
}

// startController finally starts the controller
// all error conditions MUST also be checked
// in checkController, so after a successful checkController
// startController MUST not return an error.
func (this *ControllerManager) startController(cntr Controller) error {
	for i, a := range cntr.GetDefinition().After() {
		if i == 0 {
			cntr.Infof("observing initialization requirements: %s", utils.Strings(cntr.GetDefinition().After()...))
		}
		after := this.prepared[a]
		if after != nil {
			if !after.IsReached() {
				cntr.Infof("  startup of %q waiting for %q", cntr.GetName(), a)
				if !after.Sync(this.ctx) {
					return fmt.Errorf("setup aborted")
				}
				cntr.Infof("  controller %q is initialized now", a)
			} else {
				cntr.Infof("  controller %q is already initialized", a)
			}
		} else {
			cntr.Infof("  omittimg unused controller %q", a)
		}
	}

	err := cntr.Prepare()
	if err != nil {
		return err
	}
	this.prepared[cntr.GetName()].Reach()

	ctxutil.SyncPointRunAndCancelOnExit(this.ctx, cntr.Run)
	return nil
}
