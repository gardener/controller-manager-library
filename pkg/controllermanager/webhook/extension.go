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
	"k8s.io/apimachinery/pkg/api/errors"
	"strings"

	"github.com/gardener/controller-manager-library/pkg/certs"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	parentcfg "github.com/gardener/controller-manager-library/pkg/controllermanager/config"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/extension"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/webhook/admission"
	areacfg "github.com/gardener/controller-manager-library/pkg/controllermanager/webhook/config"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"github.com/gardener/controller-manager-library/pkg/server"
	"github.com/gardener/controller-manager-library/pkg/utils"

	"k8s.io/api/admissionregistration/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
)

const TYPE = areacfg.OPTION_SOURCE

var kinds = map[string]WebhookKind{
	"MutatingWebhookConfiguration":   MUTATING,
	"ValidatingWebhookConfiguration": VALIDATING,
}

func init() {
	extension.RegisterExtension(&ExtensionType{DefaultRegistry()})
}

type ExtensionType struct {
	Registry
}

var _ extension.ExtensionType = &ExtensionType{}

func NewExtensionType() *ExtensionType {
	return &ExtensionType{NewRegistry()}
}

func (this *ExtensionType) Name() string {
	return TYPE
}

func (this *ExtensionType) Definition() extension.ExtensionDefinition {
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

func (this *ExtensionDefinition) CreateExtension(cm extension.ControllerManager) (extension.Extension, error) {
	return NewExtension(this.definitions, cm)
}

////////////////////////////////////////////////////////////////////////////////

type Extension struct {
	extension.Environment

	config         *areacfg.Config
	definitions    Definitions
	registrations  Registrations
	defaultCluster cluster.Interface
	server         *server.HTTPServer
	certificate    certs.CertificateSource
	hooks          map[string]Interface
	labels         map[string]string
}

func NewExtension(defs Definitions, cm extension.ControllerManager) (*Extension, error) {
	ext := extension.NewDefaultEnvironment(nil, TYPE, cm)
	cfg := areacfg.GetConfig(cm.GetConfig())

	if cfg.RegistrationName == "" {
		cfg.RegistrationName = cm.GetName()
	}
	if !cfg.DedicatedRegistrations {
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

func (this *Extension) getCluster(def Definition) string {
	cn := def.GetCluster()
	if cn == CLUSTER_MAIN {
		return this.config.Cluster
	}
	return cn
}

func (this *Extension) GetConfig() *areacfg.Config {
	return this.config
}

func (this *Extension) RequiredClusters() (utils.StringSet, error) {
	result := utils.StringSet{}

	for _, r := range this.registrations {
		c := this.getCluster(r)
		if c != "" {
			result.Add(c)
		}
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

		if def.GetCluster() != "" {
			if def.GetCluster() == CLUSTER_MAIN {
				target = this.defaultCluster
			} else {
				target = this.GetCluster(def.GetCluster())
				if target == nil {
					return fmt.Errorf("invalid cluster %q for webhook %q", def.GetCluster(), def.GetName())
				}
			}
		}

		w, err := NewWebhook(this, def, target)
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
			cn := this.getCluster(def)
			if cn != "" { // use unmapped cluster here with default scheme
				target := this.GetCluster(cn)
				reg := registrations.GetOrCreateGroup(target)
				if this.config.DedicatedRegistrations {
					reg.AddRegistration(def.GetName(), def.GetKind())
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
		err = this.register(registrations, this.config.RegistrationName, true)

	}
	return err
}

func (this *Extension) cleanup(cluster cluster.Interface, selector labels.Selector, keep map[string]utils.StringSet, examples ...runtime.Object) error {
	for _, example := range examples {
		r, err := cluster.Resources().GetByExample(example)
		if err != nil {
			return err
		}
		kind := r.Info().Kind()
		if err != nil {
			return err
		}

		list, err := r.List(meta.ListOptions{LabelSelector: selector.String()})
		if err != nil {
			return err
		}

		key := string(kinds[kind])

		this.Infof("found %d matching %ss  (%s)", len(list), kind, key)
		for _, found := range list {
			if !keep[found.GetName()].Contains(key) {
				this.Infof("found obsolete %s %q (%s) in cluster %q", kind, found.GetName(), keep[found.GetName()], cluster.GetName())
				err := found.Delete()
				if err != nil {
					return fmt.Errorf("cannot delete obsolete %s %q in cluster %q: %s", kind, found.GetName(), cluster.GetName(), err)
				}
			}
		}
	}
	return nil
}

func (this *Extension) register(registrations WebhookRegistrationGroups, name string, cleanup bool) error {
	for n, g := range registrations {
		if len(g.declarations) > 0 {
			this.Infof("registering webbhooks for cluster %q (%s)", n, name)
			cnt, err := CreateOrUpdateMutatingWebhookRegistration(this.labels, g.cluster, name, g.declarations...)
			if err != nil {
				return err
			}
			if cnt > 0 {
				g.AddRegistration(name, MUTATING)
				this.Infof("  found %d mutating webhooks", cnt)
			}
			cnt, err = CreateOrUpdateValidatingWebhookRegistration(this.labels, g.cluster, name, g.declarations...)
			if err != nil {
				return err
			}
			if cnt > 0 {
				g.AddRegistration(name, VALIDATING)
				this.Infof("  found %d validating webhooks", cnt)
			}
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

		if cleanup {
			err := this.cleanup(g.cluster, selector, g.registrations, &v1beta1.MutatingWebhookConfiguration{}, &v1beta1.ValidatingWebhookConfiguration{})
			if err != nil {
				return err
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
	return NewWebhookDeclaration(def.GetKind(), target, def.GetName(), def.GetNamespaces(), def.GetFailurePolicy(), client, def.GetOperations(), specs...)
}

func (this *Extension) RegisterWebhookGroup(name string, target cluster.Interface) error {
	g := this.definitions.Groups().Get(name)
	if g == nil {
		return fmt.Errorf("webhook group %q not found", name)
	}
	this.Info("registering webhook group %q for cluster %q", name, target.GetName())
	set := g.Members()
	registrations := WebhookRegistrationGroups{}
	reg := registrations.GetOrCreateGroup(target)

	grpname := this.RegistrationGroupName(name)
	for n := range set {
		w := this.hooks[n]
		if w == nil {
			this.Info("omitting inactive webhook %q", n)
			continue
		}
		def := w.GetDefinition()
		if this.config.DedicatedRegistrations {
			reg.AddRegistration(def.GetName(), def.GetKind())
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
	return this.register(registrations, grpname, false)
}

func (this *Extension) RegistrationGroupName(name string) string {
	return fmt.Sprintf("%s-%s", this.config.RegistrationName, name)
}

func (this *Extension) DeleteWebhookGroup(name string, target cluster.Interface) error {
	g := this.definitions.Groups().Get(name)
	if g == nil {
		return fmt.Errorf("webhook group %q not found", name)
	}
	this.Info("deleting webhook group %q from cluster %q", name, target.GetName())
	set := g.Members()

	for n := range set {
		w := this.hooks[n]
		if w == nil {
			continue
		}
		def := w.GetDefinition()
		if this.config.DedicatedRegistrations {
			err := this.DeleteWebhook(def, target)
			if err != nil {
				return err
			}
		}
	}
	grpname := this.RegistrationGroupName(name)
	err := DeleteMutatingWebhookRegistration(target, grpname)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	err = DeleteValidatingWebhookRegistration(target, grpname)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	return nil
}

func (this *Extension) RegisterWebhookByName(name string, target cluster.Interface) error {
	hook := this.hooks[name]
	if hook == nil {
		if this.definitions.Get(name) == nil {
			return fmt.Errorf("unknown webhook %q", name)

		}
		return fmt.Errorf("webhook %q not actice", name)
	}
	return this.RegisterWebhook(hook.GetDefinition(), target)
}

func (this *Extension) RegisterWebhook(def Definition, target cluster.Interface) error {
	wh, err := this.CreateWebhookDeclaration(def, target)
	if err != nil {
		return fmt.Errorf("webhook registration for %q failed: %s", def.GetName(), err)
	}
	this.Infof("registering %s webhook %q for cluster %q", def.GetKind(), def.GetName(), target.GetName())
	switch def.GetKind() {
	case MUTATING:
		_, err := CreateOrUpdateMutatingWebhookRegistration(this.labels, target, def.GetName(), wh)
		return err
	case VALIDATING:
		_, err := CreateOrUpdateValidatingWebhookRegistration(this.labels, target, def.GetName(), wh)
		return err
	}
	return fmt.Errorf("invalid kind %q for webhook %q", def.GetKind(), def.GetName())
}

func (this *Extension) DeleteWebhookByName(name string, target cluster.Interface) error {
	hook := this.hooks[name]
	if hook == nil {
		if this.definitions.Get(name) == nil {
			return fmt.Errorf("unknown webhook %q", name)

		}
		return fmt.Errorf("webhook %q not actice", name)
	}
	return this.DeleteWebhook(hook.GetDefinition(), target)
}

func (this *Extension) DeleteWebhook(def Definition, target cluster.Interface) error {
	this.Infof("deleting %s webhook %q for cluster %q", def.GetKind(), def.GetName(), target.GetName())
	switch def.GetKind() {
	case MUTATING:
		return DeleteMutatingWebhookRegistration(target, def.GetName())
	case VALIDATING:
		return DeleteValidatingWebhookRegistration(target, def.GetName())
	}
	return fmt.Errorf("invalid kind %q for webhook %q", def.GetKind(), def.GetName())
}
