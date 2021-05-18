/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package module

import (
	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/extension"
	areacfg "github.com/gardener/controller-manager-library/pkg/controllermanager/module/config"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/module/handler"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"github.com/gardener/controller-manager-library/pkg/utils"
)

////////////////////////////////////////////////////////////////////////////////
// Definition
////////////////////////////////////////////////////////////////////////////////

const CLUSTER_MAIN = "<MAIN>"

type HandlerType func(Interface) (handler.Interface, error)

type Environment interface {
	extension.Environment
	GetConfig() *areacfg.Config
}

type Interface interface {
	extension.ElementBase
	extension.SharedAttributes

	GetEnvironment() Environment

	GetMainCluster() cluster.Interface
	GetClusterById(id string) cluster.Interface
	GetCluster(name string) cluster.Interface
	GetClusterAliases(eff string) utils.StringSet
	GetDefinition() Definition

	GetObject(key resources.ClusterObjectKey) (resources.Object, error)
	GetCachedObject(key resources.ClusterObjectKey) (resources.Object, error)
}

type OptionDefinition extension.OptionDefinition

type Definition interface {
	Name() string
	Cluster() string
	RequiredClusters() []string
	ActivateExplicitly() bool

	ConfigOptions() extension.OptionDefinitions
	ConfigOptionSources() extension.OptionSourceDefinitions

	Handlers() map[string]HandlerType

	Definition() Definition

	String() string
}
