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
	"github.com/gardener/controller-manager-library/pkg/controllermanager/config"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/webhooks/admission"
	whcfg "github.com/gardener/controller-manager-library/pkg/controllermanager/webhooks/config"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"github.com/gardener/controller-manager-library/pkg/server"
	"github.com/gardener/controller-manager-library/pkg/utils"
)

func init() {
	controllermanager.RegisterExtension(&WebhookExtensionType{DefaultRegistry()})
}

type WebhookExtensionType struct {
	Registry
}

var _ controllermanager.ExtensionType = &WebhookExtensionType{}

func NewWebhookExtensionType() *WebhookExtensionType {
	return &WebhookExtensionType{NewRegistry()}
}

func (this *WebhookExtensionType) Name() string {
	return OPTION_SOURCE
}

func (this *WebhookExtensionType) Definition() controllermanager.ExtensionDefinition {
	return NewWebhookExtensionDefinition(this.GetDefinitions())
}

////////////////////////////////////////////////////////////////////////////////

type WebhookExtensionDefinition struct {
	definitions Definitions
}

func NewWebhookExtensionDefinition(defs Definitions) *WebhookExtensionDefinition {
	return &WebhookExtensionDefinition{
		definitions: defs,
	}
}

func (this *WebhookExtensionDefinition) Name() string {
	return OPTION_SOURCE
}

func (this *WebhookExtensionDefinition) Size() int {
	return this.definitions.Size()
}

func (this *WebhookExtensionDefinition) Names() utils.StringSet {
	return this.definitions.Names()
}

func (this *WebhookExtensionDefinition) ExtendConfig(cfg *config.Config) {
	cfg.AddSource(OPTION_SOURCE, NewConfig())
}

func (this *WebhookExtensionDefinition) CreateExtension(logctx logger.LogContext, cm *controllermanager.ControllerManager) (controllermanager.Extension, error) {
	return NewWebhookExtension(logctx, this.definitions, cm)
}

////////////////////////////////////////////////////////////////////////////////

type WebhookExtension struct {
	logger.LogContext
	config         *Config
	definitions    Definitions
	active         utils.StringSet
	manager        *controllermanager.ControllerManager
	defaultCluster cluster.Interface
	server         *HTTPServer
	certificate    certs.CertificateSource
	hooks          map[string]*admission.HTTPHandler
}

func NewWebhookExtension(logctx logger.LogContext, defs Definitions, cm *controllermanager.ControllerManager) (*WebhookExtension, error) {
	cfg := GetConfig(cm.GetConfig())

	active, err := defs.GetActiveWebhooks(cfg.Webhooks)
	if err != nil {
		return nil, err
	}
	if len(active) == 0 {
		logctx.Infof("no webhooks activated")
		return nil, nil
	}

	logctx.Infof("activated webhooks: %s", active)

	return &WebhookExtension{
		LogContext:  logctx,
		server:      NewHTTPServer(cm.GetContext(), logctx),
		config:      cfg,
		manager:     cm,
		definitions: defs,
		active:      active,
		hooks:       map[string]*admission.HTTPHandler{},
	}, nil
}

func (this *WebhookExtension) Name() string {
	return OPTION_SOURCE
}

func (this *WebhookExtension) GetContext() context.Context {
	return this.manager.GetContext()
}

func (this *WebhookExtension) RequiredClusters() (utils.StringSet, error) {
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

func (this *WebhookExtension) RegisterHandler(def Definition) error {
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

func (this *WebhookExtension) RegisterWebhook(def Definition, target cluster.Interface) error {
	var client whcfg.WebhookClientConfigSource
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
		client = whcfg.NewURLWebhookClientConfig(url, cabundle)
	} else {
		sn := resources.NewObjectName(this.manager.GetNamespace(), this.config.Service)
		if target == this.defaultCluster {
			this.Infof("registering webhook %q for cluster %s with service %s", def.GetName(), target, sn)
			client = whcfg.NewServiceWebhookClientConfig(sn, def.GetName(), cabundle)
		} else {
			this.Infof("registering webhook %q for cluster %s with runtime service %s", def.GetName(), target, sn)
			client = whcfg.NewRuntimeServiceWebhookClientConfig(sn, def.GetName(), cabundle)
		}
	}

	specs := make([]interface{}, len(def.GetResources()))
	for i, s := range def.GetResources() {
		specs[i] = s
	}

	wh, err := whcfg.NewWebhook(target, def.GetName(), def.GetNamespaces(), client, specs...)
	if err != nil {
		return err
	}
	return whcfg.CreateOrUpdateMutatingWebhookRegistration(target, def.GetName(), wh)
}

func (this *WebhookExtension) Start(ctx context.Context) error {
	var err error

	cfg := GetConfig(this.manager.GetConfig())
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
