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
	"github.com/gardener/controller-manager-library/pkg/config"
	"github.com/gardener/controller-manager-library/pkg/ctxutil"
	"github.com/gardener/controller-manager-library/pkg/server"
	"sync"
	"time"

	"github.com/gardener/controller-manager-library/pkg/configmain"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	areacfg "github.com/gardener/controller-manager-library/pkg/controllermanager/config"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/extension"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"github.com/gardener/controller-manager-library/pkg/resources/access"
	"github.com/gardener/controller-manager-library/pkg/run"
	"github.com/gardener/controller-manager-library/pkg/utils"
)

type ControllerManager struct {
	logger.LogContext
	lock       sync.Mutex
	extensions extension.Extensions

	name       string
	namespace  string
	definition *Definition

	context  context.Context
	config   *areacfg.Config
	clusters cluster.Clusters
}

var _ extension.ControllerManager = &ControllerManager{}

func NewControllerManager(ctx context.Context, def *Definition) (*ControllerManager, error) {
	maincfg := configmain.Get(ctx)
	cfg := areacfg.GetConfig(maincfg)
	lgr := logger.New()
	logger.Info("using option settings:")
	config.Print(logger.Infof, "", cfg.OptionSet)
	logger.Info("-----------------------")
	ctx = logger.Set(ctxutil.WaitGroupContext(ctx), lgr)
	ctx = context.WithValue(ctx, resources.ATTR_EVENTSOURCE, def.GetName())

	for _, e := range def.extensions {
		err := e.Validate()
		if err != nil {
			return nil, err
		}
	}

	if cfg.NamespaceRestriction {
		logger.Infof("enable namespace restriction for access control")
		access.RegisterNamespaceOnlyAccess()
	} else {
		logger.Infof("disable namespace restriction for access control")
	}

	name := def.GetName()
	if cfg.Name != "" {
		name = cfg.Name
	}

	namespace := run.GetConfig(maincfg).Namespace
	if namespace == "" {
		namespace = "kube-system"
	}

	found := false
	for _, e := range def.extensions {
		if e.Size() > 0 {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("no controller manager extension registered")
	}

	for _, e := range def.extensions {
		if e.Size() > 0 {
			logger.Infof("configured %s: %s", e.Name(), e.Names())
		}
	}

	cm := &ControllerManager{
		LogContext: lgr,
		name:       name,
		namespace:  namespace,
		definition: def,
		config:     cfg,
	}
	ctx = ctx_controllermanager.WithValue(ctx, cm)
	cm.context = ctx

	set := utils.StringSet{}
	cm.extensions = extension.Extensions{}
	for _, d := range def.extensions {
		e, err := d.CreateExtension(cm)
		if err != nil {
			return nil, err
		}
		if e == nil {
			logger.Infof("skipping unused extension %q", d.Name())
			continue
		}
		cm.extensions[d.Name()] = e
		s, err := e.RequiredClusters()
		if err != nil {
			return nil, err
		}
		set.AddSet(s)
	}

	if len(cm.extensions) == 0 {
		return nil, fmt.Errorf("no controller manager extension activated")
	}

	clusters, err := def.ClusterDefinitions().CreateClusters(ctx, lgr, cfg, set)
	if err != nil {
		return nil, err
	}

	cm.clusters = clusters

	return cm, nil
}

func (this *ControllerManager) GetName() string {
	return this.name
}

func (this *ControllerManager) GetNamespace() string {
	return this.namespace
}

func (this *ControllerManager) GetContext() context.Context {
	return this.context
}

func (this *ControllerManager) GetConfig() *areacfg.Config {
	return this.config
}

func (this *ControllerManager) GetExtension(name string) extension.Extension {
	return this.extensions[name]
}

func (this *ControllerManager) ClusterDefinitions() cluster.Definitions {
	return this.definition.ClusterDefinitions()
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

	server.ServeFromMainConfig(this.context, "httpserver")

	for _, e := range this.extensions {
		err = e.Start(this.context)
		if err != nil {
			return err
		}
	}

	<-this.context.Done()
	this.Info("waiting for extensions to shutdown")
	ctxutil.WaitGroupWait(this.context, 120*time.Second)
	this.Info("all extensions down -> exit controller manager")
	return nil
}
