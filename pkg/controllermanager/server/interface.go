/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package server

import (
	"net/http"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/extension"
	areacfg "github.com/gardener/controller-manager-library/pkg/controllermanager/server/config"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/server/handler"
	"github.com/gardener/controller-manager-library/pkg/server"
)

////////////////////////////////////////////////////////////////////////////////
// Definition
////////////////////////////////////////////////////////////////////////////////

const CLUSTER_MAIN = "<MAIN>"

type ServerKind string

const HTTP = ServerKind("http")
const HTTPS = ServerKind("https")

type HandlerType func(Interface) (handler.Interface, error)

type Environment interface {
	extension.Environment
	GetConfig() *areacfg.Config
}

type Interface interface {
	extension.ElementBase
	extension.SharedAttributes

	GetEnvironment() Environment
	GetDefinition() Definition
	GetCluster() cluster.Interface

	GetKind() ServerKind

	Server() *server.HTTPServer

	Register(pattern string, handler http.HandlerFunc)
	RegisterHandler(pattern string, handler http.Handler)
}

type OptionDefinition extension.OptionDefinition

type Definition interface {
	Name() string
	Kind() ServerKind
	Cluster() string
	ActivateExplicitly() bool

	ConfigOptions() extension.OptionDefinitions
	ConfigOptionSources() extension.OptionSourceDefinitions

	ServerPort() int

	Handlers() map[string]HandlerType

	Definition() Definition
}
