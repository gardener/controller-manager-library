/*
 * Copyright 2019 SAP SE or an SAP affiliate company. All rights reserved.
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

package webhooks

import (
	"context"
	"fmt"

	"github.com/gardener/controller-manager-library/pkg/certs"
	"github.com/gardener/controller-manager-library/pkg/controllermanager"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	parentcfg "github.com/gardener/controller-manager-library/pkg/controllermanager/config"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/webhooks/admission"
	areacfg "github.com/gardener/controller-manager-library/pkg/controllermanager/webhooks/config"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"github.com/gardener/controller-manager-library/pkg/server"
	"github.com/gardener/controller-manager-library/pkg/utils"
)

const TYPE = areacfg.OPTION_SOURCE

func init() {
	controllermanager.RegisterExtension(&ExtensionType{DefaultRegistry()})
}

type ExtensionType struct {
	Registry
}

var _ controllermanager.ExtensionType = &ExtensionType{}

func NewExtensionType() *ExtensionType {
	return &ExtensionType{NewRegistry()}
}

func (this *ExtensionType) Name() string {
	return TYPE
}

func (this *ExtensionType) Definition() controllermanager.ExtensionDefinition {
	return NewExtensionDefinition(this.GetDefinitions())
}

////////////////////////////////////////////////////////////////////////////////

type ExtensionDefinition struct {
	definitions Definitions
}

func NewExtensionDefinition(defs Definitions) *ExtensionDefinition {
	return &ExtensionDefinition{
		definitions: defs,
	}
}

func (this *ExtensionDefinition) Name() string {
	return TYPE
}

func (this *ExtensionDefinition) Size() int {
	return this.definitions.Size()
}

func (this *ExtensionDefinition) Names() utils.StringSet {
	return this.definitions.Names()
}

func (this *ExtensionDefinition) Validate() error {
	return nil
}

func (this *ExtensionDefinition) ExtendConfig(cfg *parentcfg.Config) {
	cfg.AddSource(areacfg.OPTION_SOURCE, areacfg.NewConfig())
}

func (this *ExtensionDefinition) CreateExtension(logctx logger.LogContext, cm *controllermanager.ControllerManager) (controllermanager.Extension, error) {
	return NewExtension(logctx, this.definitions, cm)
}

////////////////////////////////////////////////////////////////////////////////

type Extension struct {
	logger.LogContext
	config         *areacfg.Config
	definitions    Definitions
	registrations  Registrations
	manager        *controllermanager.ControllerManager
	defaultCluster cluster.Interface
	server         *HTTPServer
	certificate    certs.CertificateSource
	hooks          map[string]*admission.HTTPHandler
}

func NewExtension(logctx logger.LogContext, defs Definitions, cm *controllermanager.ControllerManager) (*Extension, error) {
	cfg := areacfg.GetConfig(cm.GetConfig())

	active, err := defs.GetActiveWebhooks(cfg.Webhooks)
	if err != nil {
		return nil, err
	}
	if len(active) == 0 {
		logctx.Infof("no webhooks activated")
		return nil, nil
	}

	registrations, err := defs.Registrations(active.AsArray()...)
	if err != nil {
		return nil, err
	}
	logctx.Infof("activated webhooks: %s", active)

	return &Extension{
		LogContext:    logctx,
		server:        NewHTTPServer(cm.GetContext(), logctx),
		config:        cfg,
		manager:       cm,
		definitions:   defs,
		registrations: registrations,
		hooks:         map[string]*admission.HTTPHandler{},
	}, nil
}

func (this *Extension) Name() string {
	return TYPE
}

func (this *Extension) GetContext() context.Context {
	return this.manager.GetContext()
}

func (this *Extension) RequiredClusters() (utils.StringSet, error) {
	result := utils.StringSet{}

	defs := this.definitions
	active, err := defs.GetActiveWebhooks(this.config.Webhooks)
	if err != nil {
		return nil, err
	}

	for n := range active {
		r := defs.Get(n)
		c := r.GetCluster()
		result.Add(c)
	}
	result.Add(this.config.Cluster)
	return result, nil
}

func (this *Extension) RegisterHandler(def Definition) error {
	if this.hooks[def.GetName()] != nil {
		return fmt.Errorf("handler for webhook with name %q already registed", def.GetName())
	}
	handler, err := def.GetHandlerType()(this)
	if err != nil {
		return err
	}
	var target cluster.Interface
	if def.GetCluster() == "" {
		target = this.defaultCluster
	} else {
		target = this.manager.GetCluster(def.GetCluster())
	}
	http := admission.New(this.NewContext("webhook", def.GetName()), target.ResourceContext().Scheme(), handler)
	this.hooks[def.GetName()] = http
	server.RegisterHandler(def.GetName(), http)
	return nil
}

func (this *Extension) RegisterWebhook(def Definition, target cluster.Interface) error {
	var client WebhookClientConfigSource
	cabundle := this.certificate.GetCertificateInfo().CACert()
	if len(cabundle) == 0 {
		return fmt.Errorf("no cert authority given")
	}
	if this.config.Hostname != "" {
		url := fmt.Sprintf("https://%s/%s", this.config.Hostname, def.GetName())
		if this.config.Port > 0 {
			url = fmt.Sprintf("https://%s:%d/%s", this.config.Hostname, this.config.Port, def.GetName())
		}
		this.Infof("registering webhook %q for cluster %s with URL %s", def.GetName(), target, url)
		client = NewURLWebhookClientConfig(url, cabundle)
	} else {
		sn := resources.NewObjectName(this.manager.GetNamespace(), this.config.Service)
		if target == this.defaultCluster {
			this.Infof("registering webhook %q for cluster %s with service %s", def.GetName(), target, sn)
			client = NewServiceWebhookClientConfig(sn, def.GetName(), cabundle)
		} else {
			this.Infof("registering webhook %q for cluster %s with runtime service %s", def.GetName(), target, sn)
			client = NewRuntimeServiceWebhookClientConfig(sn, def.GetName(), cabundle)
		}
	}

	specs := make([]interface{}, len(def.GetResources()))
	for i, s := range def.GetResources() {
		specs[i] = s
	}

	wh, err := NewWebhook(target, def.GetName(), def.GetNamespaces(), client, specs...)
	if err != nil {
		return err
	}
	return CreateOrUpdateMutatingWebhookRegistration(target, def.GetName(), wh)
}

func (this *Extension) Start(ctx context.Context) error {
	var err error

	cfg := areacfg.GetConfig(this.manager.GetConfig())
	this.defaultCluster = this.manager.GetCluster(cfg.Cluster)

	if this.defaultCluster == nil {
		return fmt.Errorf("default cluster %q for webhook server not found", cfg.Cluster)
	}

	if cfg.CertFile != "" {
		this.certificate, err = CreateFileCertificateSource(ctx, this, cfg)
	} else {
		this.certificate, err = CreateSecretCertificateSource(ctx, this, cfg, this.manager)
	}
	if err != nil {
		return err
	}

	defs := this.definitions
	active, err := defs.GetActiveWebhooks(cfg.Webhooks)
	if err != nil {
		return err
	}
	if len(active) == 0 {
		this.Infof("no webhooks activated")
		return nil
	}

	for n := range active {
		this.RegisterHandler(defs.Get(n))
	}
	if !this.config.OmitRegistrations {
		for n := range active {
			def := this.definitions.Get(n)
			cn := def.GetCluster()
			if cn != "" {
				target := this.manager.GetCluster(cn)
				err := this.RegisterWebhook(def, target)
				if err != nil {
					return err
				}
			}
		}
	}
	this.server.Start(this.certificate, "", this.config.Port)
	return nil
}
