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

package webhook

import (
	"context"
	"fmt"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller/mappings"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/webhook/admission"
	"k8s.io/apimachinery/pkg/runtime"
	"strings"

	"github.com/gardener/controller-manager-library/pkg/certs"
	"github.com/gardener/controller-manager-library/pkg/controllermanager"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	parentcfg "github.com/gardener/controller-manager-library/pkg/controllermanager/config"
	areacfg "github.com/gardener/controller-manager-library/pkg/controllermanager/webhook/config"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"github.com/gardener/controller-manager-library/pkg/server"
	"github.com/gardener/controller-manager-library/pkg/utils"

	"k8s.io/api/admissionregistration/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
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
	ecfg := areacfg.NewConfig()
	this.definitions.ExtendConfig(ecfg)
	cfg.AddSource(areacfg.OPTION_SOURCE, ecfg)
}

func (this *ExtensionDefinition) CreateExtension(cm *controllermanager.ControllerManager) (controllermanager.Extension, error) {
	return NewExtension(this.definitions, cm)
}

////////////////////////////////////////////////////////////////////////////////

type Extension struct {
	controllermanager.Environment

	config         *areacfg.Config
	definitions    Definitions
	registrations  Registrations
	defaultCluster cluster.Interface
	server         *server.HTTPServer
	certificate    certs.CertificateSource
	hooks          map[string]Interface
	labels         map[string]string
}

func NewExtension(defs Definitions, cm *controllermanager.ControllerManager) (*Extension, error) {
	ext := controllermanager.NewDefaultEnvironment(nil, TYPE, cm)
	cfg := areacfg.GetConfig(cm.GetConfig())

	if !cfg.DedicatedRegistrations {
		if cfg.RegistrationName == "" {
			cfg.RegistrationName = cm.GetName()
		}
		ext.Infof("using grouped webhook registrations per cluster with name %q", cfg.RegistrationName)
	}

	groups := defs.Groups()
	ext.Infof("configured groups: %s", groups.AllGroups())

	active, err := groups.Activate(ext, strings.Split(cfg.Webhooks, ","))
	if err != nil {
		return nil, err
	}
	if len(active) == 0 {
		ext.Infof("no webhooks activated")
		return nil, nil
	}

	registrations, err := defs.Registrations(active.AsArray()...)
	if err != nil {
		return nil, err
	}

	spec := cfg.Service + "--" + cm.GetNamespace()
	if cfg.Hostname != "" {
		spec = cfg.Hostname
	}
	labels := map[string]string{
		"service": spec,
	}
	for _, l := range cfg.Labels {
		a := strings.Split(l, "=")
		labels[a[0]] = a[1]
	}
	return &Extension{
		Environment:   ext,
		server:        server.NewHTTPServer(ext.GetContext(), ext, "webhook"),
		config:        cfg,
		definitions:   defs,
		registrations: registrations,
		hooks:         map[string]Interface{},
		labels:        labels,
	}, nil
}

func (this *Extension) GetConfig() *areacfg.Config {
	return this.config
}

func (this *Extension) RequiredClusters() (utils.StringSet, error) {
	result := utils.StringSet{}

	for _, r := range this.registrations {
		c := r.GetCluster()
		result.Add(c)
	}
	result.Add(this.config.Cluster)
	return result, nil
}

func (this *Extension) Start(ctx context.Context) error {
	var err error

	this.defaultCluster = this.GetCluster(this.config.Cluster)

	if this.defaultCluster == nil {
		return fmt.Errorf("default cluster %q for webhook server not found", this.config.Cluster)
	}

	if this.config.CertFile != "" {
		this.certificate, err = CreateFileCertificateSource(ctx, this)
	} else {
		this.certificate, err = CreateSecretCertificateSource(ctx, this)
	}
	if err != nil {
		return err
	}

	if len(this.registrations) == 0 {
		this.Infof("no webhooks activated")
		return nil
	}

	for _, def := range this.registrations {
		var target cluster.Interface
		var scheme *runtime.Scheme

		if def.GetCluster() != "" {
			if def.GetCluster() == mappings.CLUSTER_MAIN {
				target = this.defaultCluster
			} else {
				target = this.GetCluster(def.GetCluster())
				if target == nil {
					return fmt.Errorf("invalid cluster %q for webhook %q", def.GetCluster(), def.GetName())
				}
			}
			scheme = target.ResourceContext().Scheme()
		}

		w, err := NewWebhook(this, def, scheme, target)
		if err != nil {
			return err
		}
		this.RegisterHandler(w)
	}
	this.server.Start(this.certificate, "", this.config.Port)

	if !this.config.OmitRegistrations {
		registrations := WebhookRegistrationGroups{}

		for _, w := range this.hooks {
			def := w.GetDefinition()
			cn := def.GetCluster()
			if cn != "" {
				target := this.GetCluster(cn)
				reg := registrations.Assure(target)
				if this.config.DedicatedRegistrations {
					reg.AddRegistration(def.GetName())
					err := this.RegisterWebhook(def, target)
					if err != nil {
						return err
					}
				} else {
					wh, err := this.CreateWebhookDeclaration(def, target)
					if err != nil {
						return fmt.Errorf("webhook registration for %q failed: %s", def.GetName(), err)
					}
					reg.AddDeclaration(wh)
				}
			}
		}
		for n, g := range registrations {
			if len(g.declarations) > 0 {
				this.Infof("registering webbhooks for cluster %q (%s)", n, this.config.RegistrationName)
				g.AddRegistration(this.config.RegistrationName)
				err := CreateOrUpdateMutatingWebhookRegistration(this.labels, g.cluster, this.config.RegistrationName, g.declarations...)
				if err != nil {
					return err
				}
			}
			r, err := g.cluster.Resources().GetByExample(&v1beta1.MutatingWebhookConfiguration{})
			if err != nil {
				return err
			}
			selector := labels.NewSelector()
			for k, v := range this.labels {
				r, err := labels.NewRequirement(k, selection.Equals, []string{v})
				if err != nil {
					return err
				}
				selector = selector.Add(*r)
			}
			this.Infof("looking for obsolete registrations: %s", selector.String())

			list, err := r.List(meta.ListOptions{LabelSelector: selector.String()})
			if err != nil {
				return err
			}
			for _, found := range list {
				if !g.registrations.Contains(found.GetName()) {
					this.Infof("found obsolete registration %q", found.GetName())
					err := found.Delete()
					if err != nil {
						return fmt.Errorf("cannot delete obsolete webhook registration %q: %s", found.GetName(), err)
					}
				}
			}
		}
	}
	return nil
}

func (this *Extension) RegisterHandler(wh Interface) error {
	if this.hooks[wh.GetName()] != nil {
		return fmt.Errorf("handler for webhook with name %q already registed", wh.GetName())
	}
	this.hooks[wh.GetName()] = wh
	this.server.RegisterHandler(wh.GetName(), admission.New(wh, wh.GetScheme(), wh))
	return nil
}

func (this *Extension) CreateWebhookDeclaration(def Definition, target cluster.Interface) (*WebhookDeclaration, error) {
	var client WebhookClientConfigSource
	cabundle := this.certificate.GetCertificateInfo().CACert()
	if len(cabundle) == 0 {
		return nil, fmt.Errorf("no cert authority given")
	}
	if this.config.Hostname != "" {
		if target == this.defaultCluster && this.config.Service != "" {
			sn := resources.NewObjectName(this.Namespace(), this.config.Service)
			this.Infof("registering webhook %q for cluster %q with service %q", def.GetName(), target, sn)
			client = NewServiceWebhookClientConfig(sn, def.GetName(), cabundle)
		} else {
			url := fmt.Sprintf("https://%s/%s", this.config.Hostname, def.GetName())
			if this.config.Port > 0 {
				url = fmt.Sprintf("https://%s:%d/%s", this.config.Hostname, this.config.Port, def.GetName())
			}
			this.Infof("registering webhook %q for cluster %q with URL %q", def.GetName(), target, url)
			client = NewURLWebhookClientConfig(url, cabundle)
		}
	} else {
		sn := resources.NewObjectName(this.Namespace(), this.config.Service)
		if target == this.defaultCluster {
			this.Infof("registering webhook %q for cluster %q with service %q", def.GetName(), target, sn)
			client = NewServiceWebhookClientConfig(sn, def.GetName(), cabundle)
		} else {
			this.Infof("registering webhook %q for cluster %q with runtime service %q", def.GetName(), target, sn)
			client = NewRuntimeServiceWebhookClientConfig(sn, def.GetName(), cabundle)
		}
	}

	specs := make([]interface{}, len(def.GetResources()))
	for i, s := range def.GetResources() {
		specs[i] = s
	}
	return NewWebhookDeclaration(target, def.GetName(), def.GetNamespaces(), def.GetFailurePolicy(), client, def.GetOperations(), specs...)
}

func (this *Extension) RegisterWebhook(def Definition, target cluster.Interface) error {
	wh, err := this.CreateWebhookDeclaration(def, target)
	if err != nil {
		return fmt.Errorf("webhook registration for %q failed: %s", def.GetName(), err)
	}
	return CreateOrUpdateMutatingWebhookRegistration(this.labels, target, def.GetName(), wh)
}
