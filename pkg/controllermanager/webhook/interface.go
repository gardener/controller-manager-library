/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package webhook

import (
	"net/http"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/gardener/controller-manager-library/pkg/config"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/extension"
	areacfg "github.com/gardener/controller-manager-library/pkg/controllermanager/webhook/config"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"github.com/gardener/controller-manager-library/pkg/resources/apiextensions"
)

////////////////////////////////////////////////////////////////////////////////
// Definition
////////////////////////////////////////////////////////////////////////////////

const CLUSTER_MAIN = "<MAIN>"

type WebhookKind string

const MUTATING = WebhookKind("mutating")
const VALIDATING = WebhookKind("validating")
const CONVERTING = WebhookKind("converting")

type Environment interface {
	extension.Environment
	GetConfig() *areacfg.Config

	CreateWebhookClientConfig(msg string, def Definition, target resources.Cluster) (apiextensions.WebhookClientConfigSource, error)

	RegisterWebhookByName(name string, target cluster.Interface, client apiextensions.WebhookClientConfigSource) error
	RegisterWebhook(def Definition, target cluster.Interface, client apiextensions.WebhookClientConfigSource) error
	RegisterWebhookGroup(name string, target cluster.Interface, client apiextensions.WebhookClientConfigSource) error

	DeleteWebhookByName(name string, target cluster.Interface) error
	DeleteWebhook(def Definition, target cluster.Interface) error
}

type Interface interface {
	extension.ElementBase
	//admission.Interface

	GetEnvironment() Environment
	GetDefinition() Definition
	GetCluster() cluster.Interface
	GetScheme() *runtime.Scheme
	GetKind() WebhookKind
	GetKindConfig() config.OptionSource
}

type OptionDefinition extension.OptionDefinition

type Definition interface {
	Name() string
	Resources() []extension.ResourceKey
	Scheme() *runtime.Scheme
	Kind() WebhookKind
	Handler() WebhookHandler
	Cluster() string
	ActivateExplicitly() bool

	ConfigOptions() map[string]OptionDefinition
	ConfigOptionSources() extension.OptionSourceDefinitions

	Definition() Definition
}

type WebhookHandler interface {
	GetKind() WebhookKind
	GetHTTPHandler(wh Interface) (http.Handler, error)

	String() string
}

type WebhookValidator interface {
	Validate(Interface) error
}

type HandlerFactory interface {
	CreateHandler() WebhookHandler
}

type RegistrationContext interface {
	logger.LogContext
	Maintainer() extension.MaintainerInfo
	Config() config.OptionSource
}

type RegistrationHandler interface {
	Kind() WebhookKind
	OptionSourceCreator() extension.OptionSourceCreator
	RequireDedicatedRegistrations() bool
	RegistrationNames(def Definition) []string
	RegistrationResource() runtime.Object
	CreateDeclarations(log logger.LogContext, def Definition, target cluster.Interface, client apiextensions.WebhookClientConfigSource) (WebhookDeclarations, error)
	Register(ctx RegistrationContext, labels map[string]string, cluster cluster.Interface, name string, declaration ...WebhookDeclaration) error
	Delete(log logger.LogContext, name string, def Definition, cluster cluster.Interface) error
}

type WebhookDeclaration interface {
	Kind() WebhookKind
}

type WebhookDeclarations []WebhookDeclaration

type ValidatorFunc func(Interface) error

func (this ValidatorFunc) Validate(wh Interface) error {
	return this(wh)
}

////////////////////////////////////////////////////////////////////////////////

// WebhookKindHandlerProvider is registered for a dedicated kind and
// is called to create a dedicated kind handler for a dedicated kind used in a dedicated
// Webhook etension instance
type WebhookKindHandlerProvider func(Environment, WebhookKind) (WebhookKindHandler, error)

// WebhookKindHandler gets called by a webhook extension for every instance
// of a dedicated webhook kind it is created for by a WebhookKindHandlerProvider
type WebhookKindHandler interface {
	Register(Interface) error
}
