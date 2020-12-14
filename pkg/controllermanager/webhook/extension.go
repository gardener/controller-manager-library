/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package webhook

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/gardener/controller-manager-library/pkg/certmgmt"
	"github.com/gardener/controller-manager-library/pkg/certs"
	"github.com/gardener/controller-manager-library/pkg/config"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	parentcfg "github.com/gardener/controller-manager-library/pkg/controllermanager/config"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/extension"
	areacfg "github.com/gardener/controller-manager-library/pkg/controllermanager/webhook/config"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"github.com/gardener/controller-manager-library/pkg/resources/apiextensions"
	"github.com/gardener/controller-manager-library/pkg/server"
	"github.com/gardener/controller-manager-library/pkg/utils"
)

const TYPE = areacfg.OPTION_SOURCE

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

func (this *ExtensionType) Definition() extension.Definition {
	return NewExtensionDefinition(this.GetDefinitions())
}

////////////////////////////////////////////////////////////////////////////////

type ExtensionDefinition struct {
	extension.ExtensionDefinitionBase
	definitions Definitions
}

func NewExtensionDefinition(defs Definitions) *ExtensionDefinition {
	return &ExtensionDefinition{
		ExtensionDefinitionBase: extension.NewExtensionDefinitionBase(TYPE, []string{"modules"}),
		definitions:             defs,
	}
}

func (this *ExtensionDefinition) Description() string {
	kinds := utils.StringSet{}
	for n := range this.definitions.Names() {
		kinds.Add(string(this.definitions.Get(n).Kind()))
	}
	return "webhook extension " + kinds.String()
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
	regctxs        map[WebhookKind]RegistrationContext
	definitions    Definitions
	registrations  Registrations
	defaultCluster cluster.Interface
	server         *server.HTTPServer
	certificate    certs.CertificateSource
	hooks          map[string]Interface
	labels         map[string]string
	kindHandlers   map[WebhookKind]WebhookKindHandler
	maintained     MaintainedRegistrations
}

type registrationContext struct {
	logger.LogContext
	ext    *Extension
	config config.OptionSource
}

func (this *registrationContext) Maintainer() extension.MaintainerInfo {
	return this.ext.Maintainer()
}

func (this *registrationContext) Config() config.OptionSource {
	return this.config
}

var _ RegistrationContext = (*registrationContext)(nil)

func NewExtension(defs Definitions, cm extension.ControllerManager) (*Extension, error) {
	ext := extension.NewDefaultEnvironment(nil, TYPE, cm)
	cfg := areacfg.GetConfig(cm.GetConfig())

	if cfg.RegistrationName == "" {
		cfg.RegistrationName = cm.GetName()
	}
	if !cfg.DedicatedRegistrations {
		ext.Infof("using grouped webhook registrations per cluster with name %q", cfg.RegistrationName)
	}

	cfg.CommonName = cm.GetName()
	cfg.Organization = "kubernetes"

	groups := defs.Groups()
	ext.Infof("configured groups: %s", groups.AllGroups())

	active, err := groups.Members(ext, strings.Split(cfg.Webhooks, ","))
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
	if len(cfg.Hostnames) > 0 {
		spec = cfg.Hostnames[0]
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

func (this *Extension) Maintainer() extension.MaintainerInfo {
	return this.ControllerManager().GetMaintainer()
}

func (this *Extension) getCluster(def Definition) string {
	cn := def.Cluster()
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

func (this *Extension) RequiredClusterIds(clusters cluster.Clusters) utils.StringSet {
	return nil
}

func (this *Extension) Setup(ctx context.Context) error {
	var err error

	ctxs := map[WebhookKind]RegistrationContext{}
	for _, r := range this.registrations {
		ctx := &registrationContext{this, this, nil}
		s := this.config.GetSource(string(r.Kind()))
		if s != nil {
			ctx.config = s
		}
		ctxs[r.Kind()] = ctx
	}
	this.regctxs = ctxs
	this.defaultCluster = this.GetCluster(this.config.Cluster)
	if this.defaultCluster == nil {
		return fmt.Errorf("default cluster %q for webhook server not found", this.config.Cluster)
	}

	this.kindHandlers, err = createKindHandlers(this)
	if err != nil {
		return err
	}
	for _, def := range this.registrations {
		var target cluster.Interface

		if def.Cluster() != "" {
			if def.Cluster() == CLUSTER_MAIN {
				target = this.defaultCluster
			} else {
				target = this.GetCluster(def.Cluster())
				if target == nil {
					return fmt.Errorf("invalid cluster %q for webhook %q", def.Cluster(), def.Name())
				}
			}
		}

		w, err := NewWebhook(this, def, target)
		if err != nil {
			return err
		}

		this.RegisterHandler(w)
		if kh := this.kindHandlers[def.Kind()]; kh != nil {
			err := kh.Register(w)
			if err != nil {
				return err
			}
		} else {
			this.Infof("no handler for %s(%s)", w.GetName(), def.Kind())
		}
	}
	return nil
}

func (this *Extension) certificateUpdated() {
	this.Infof("server certificate for web hooks updated")
	this.maintained.TriggerRegistrationUpdate(this)
}

func (this *Extension) Start(ctx context.Context) error {
	var err error

	this.certificate, err = this.config.CertConfig.CreateAccess(ctx, this, this.GetCluster(this.config.Cluster), this.Namespace())
	if err != nil {
		return err
	}

	if w, ok := this.certificate.(certs.Watchable); ok {
		this.Infof("server certificate is watchable -> register change notification")
		w.RegisterConsumer(certs.CertificateUpdaterFunc(func(info certmgmt.CertificateInfo) {
			this.certificateUpdated()
		}))
	}
	if len(this.registrations) == 0 {
		this.Infof("no webhooks activated")
		return nil
	}

	this.server.Start(this.certificate, "", this.config.Port)

	if !this.config.OmitRegistrations {
		registrations := WebhookRegistrationGroups{}

		for _, w := range this.hooks {
			def := w.GetDefinition()
			handler := GetRegistrationHandler(def.Kind())
			if handler == nil {
				this.Infof("no registrations for %s(%s)", w.GetName(), w.GetKind())
				continue
			}
			cn := this.getCluster(def)
			if cn != "" { // use unmapped cluster here with default scheme
				this.Infof("handle registration of %s(%s) on cluster %s", w.GetName(), w.GetKind(), cn)
				target := this.GetCluster(cn)
				reg := registrations.GetOrCreateGroup(target)
				client, err := this.CreateWebhookClientConfig("using ", def, target)
				if err != nil {
					return err
				}
				err = this.addHook(w, target, client, reg)
				if err != nil {
					return err
				}
			} else {
				this.Infof("no cluster for registration of %s(%s)", w.GetName(), w.GetKind(), cn)
			}
		}
		err = this.handleRegistrationGroups(registrations, this.config.RegistrationName, true)

	} else {
		this.Infof("omit registrations")
	}
	return err
}

func (this *Extension) RegisterHandler(wh Interface) error {
	if this.hooks[wh.GetName()] != nil {
		return fmt.Errorf("handler for webhook with name %q already registed", wh.GetName())
	}
	h, err := wh.GetDefinition().Handler().GetHTTPHandler(wh)
	if err != nil {
		return err
	}
	this.hooks[wh.GetName()] = wh
	this.server.RegisterHandler(wh.GetName(), h)
	return nil
}

func (this *Extension) CreateWebhookClientConfig(msg string, def Definition, target resources.Cluster) (apiextensions.WebhookClientConfigSource, error) {
	var client apiextensions.WebhookClientConfigSource
	cabundle := this.certificate.GetCertificateInfo().CACert()
	if len(cabundle) == 0 {
		return nil, fmt.Errorf("no cert authority given")
	}
	if msg != "" && !strings.HasPrefix(msg, " ") {
		msg = msg + " "
	}
	if len(this.config.Hostnames) > 0 {
		if target == this.defaultCluster && this.config.Service != "" {
			sn := resources.NewObjectName(this.Namespace(), this.config.Service)
			this.Infof("%swebhook %q for cluster %q with service %q", msg, def.Name(), target, sn)
			client = apiextensions.NewServiceWebhookClientConfig(sn, this.config.ServicePort, def.Name(), cabundle)
		} else {
			url := fmt.Sprintf("https://%s/%s", this.config.Hostnames[0], def.Name())
			if this.config.Port > 0 {
				url = fmt.Sprintf("https://%s:%d/%s", this.config.Hostnames[0], this.config.Port, def.Name())
			}
			this.Infof("%swebhook %q for cluster %q with URL %q", msg, def.Name(), target, url)
			client = apiextensions.NewURLWebhookClientConfig(url, cabundle)
		}
	} else {
		sn := resources.NewObjectName(this.Namespace(), this.config.Service)
		if target == this.defaultCluster {
			this.Infof("%swebhook %q for cluster %q with service %q", msg, def.Name(), target, sn)
			client = apiextensions.NewServiceWebhookClientConfig(sn, this.config.ServicePort, def.Name(), cabundle)
		} else {
			this.Infof("%swebhook %q for cluster %q with runtime service %q", msg, def.Name(), target, sn)
			client = apiextensions.NewRuntimeServiceWebhookClientConfig(sn, def.Name(), cabundle)
		}
	}
	return client, nil
}

func (this *Extension) RegisterWebhookGroup(name string, target cluster.Interface, client apiextensions.WebhookClientConfigSource) error {
	var err error

	g := this.definitions.Groups().Get(name)
	if g == nil {
		return fmt.Errorf("webhook group %q not found", name)
	}
	this.Infof("registering webhook group %q for cluster %q", name, target.GetName())
	set := g.Members()
	registrations := WebhookRegistrationGroups{}
	reg := registrations.GetOrCreateGroup(target)
	grpname := this.RegistrationGroupName(name)
	for n := range set {
		w := this.hooks[n]
		if w == nil {
			this.Infof("omitting inactive webhook %q", n)
			continue
		}
		if client == nil {
			client, err = this.CreateWebhookClientConfig("using ", w.GetDefinition(), target)
			if err != nil {
				return err
			}
		}
		this.addHook(w, target, client, reg)
	}
	return this.handleRegistrationGroups(registrations, grpname, false)
}

func (this *Extension) addHook(w Interface, target cluster.Interface, client apiextensions.WebhookClientConfigSource, reg *WebhookRegistrationGroup) error {
	def := w.GetDefinition()
	if handler := GetRegistrationHandler(def.Kind()); handler != nil {
		if this.config.DedicatedRegistrations || handler.RequireDedicatedRegistrations() {
			reg.AddRegistrations(def.Kind(), handler.RegistrationNames(def)...)
			err := this.RegisterWebhook(def, target, client)
			if err != nil {
				return err
			}
		} else {
			decls, err := handler.CreateDeclarations(this, def, target, client)
			if err != nil {
				return fmt.Errorf("webhook registration for %q failed: %s", def.Name(), err)
			}
			reg.AddDeclarations(decls...)
		}
	}
	return nil
}

func (this *Extension) RegistrationGroupName(name string) string {
	return fmt.Sprintf("%s-%s", this.config.RegistrationName, name)
}

func (this *Extension) DeleteWebhookGroup(name string, target cluster.Interface) error {
	g := this.definitions.Groups().Get(name)
	if g == nil {
		return fmt.Errorf("webhook group %q not found", name)
	}
	this.Infof("deleting webhook group %q from cluster %q", name, target.GetName())
	set := g.Members()
	kinds := map[WebhookKind]RegistrationHandler{}

	for n := range set {
		w := this.hooks[n]
		if w == nil {
			continue
		}
		def := w.GetDefinition()
		if handler := GetRegistrationHandler(def.Kind()); handler != nil {
			if this.config.DedicatedRegistrations || handler.RequireDedicatedRegistrations() {
				err := this.removeRegistration(handler, def.Name(), def, target)
				if err != nil {
					return err
				}
			}
			kinds[def.Kind()] = handler
		}
	}
	grpname := this.RegistrationGroupName(name)
	for _, handler := range kinds {
		err := this.removeRegistration(handler, grpname, nil, target)
		if err != nil && !errors.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (this *Extension) RegisterWebhookByName(name string, target cluster.Interface, client apiextensions.WebhookClientConfigSource) error {
	hook := this.hooks[name]
	if hook == nil {
		if this.definitions.Get(name) == nil {
			return fmt.Errorf("unknown webhook %q", name)

		}
		return fmt.Errorf("webhook %q not actice", name)
	}
	return this.RegisterWebhook(hook.GetDefinition(), target, client)
}

func (this *Extension) RegisterWebhook(def Definition, target cluster.Interface, client apiextensions.WebhookClientConfigSource) error {
	var err error

	handler := GetRegistrationHandler(def.Kind())
	if handler == nil {
		return fmt.Errorf("reregistrations for kind %q for webhook %q not possible", def.Kind(), def.Name())
	}
	if client == nil {
		client, err = this.CreateWebhookClientConfig("using ", def, target)
		if err != nil {
			return err
		}
	}
	decls, err := handler.CreateDeclarations(this, def, target, client)
	if err != nil {
		return fmt.Errorf("webhook registration for %q failed: %s", def.Name(), err)
	}
	return this.addRegistration(handler, def.Name(), def, target, decls...)
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
	handler := GetRegistrationHandler(def.Kind())
	if handler != nil {
		this.removeRegistration(handler, def.Name(), def, target)
	}
	return nil
}
