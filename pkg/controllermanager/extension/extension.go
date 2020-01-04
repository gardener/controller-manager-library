/*
 * Copyright 2020 SAP SE or an SAP affiliate company. All rights reserved.
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

package extension

import (
	"context"
	"fmt"
	"github.com/gardener/controller-manager-library/pkg/config"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	"github.com/gardener/controller-manager-library/pkg/ctxutil"
	"sync"
	"time"

	areacfg "github.com/gardener/controller-manager-library/pkg/controllermanager/config"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/utils"
)

type ExtensionDefinitions map[string]ExtensionDefinition
type ExtensionTypes map[string]ExtensionType
type Extensions map[string]Extension

type ExtensionType interface {
	Name() string
	Definition() ExtensionDefinition
}

type ExtensionDefinition interface {
	Name() string
	Names() utils.StringSet
	Size() int
	Validate() error
	ExtendConfig(*areacfg.Config)
	CreateExtension(cm ControllerManager) (Extension, error)
}

type Extension interface {
	Name() string
	RequiredClusters() (utils.StringSet, error)
	Start(ctx context.Context) error
}

type ExtensionRegistry interface {
	RegisterExtension(e ExtensionType) error
	MustRegisterExtension(e ExtensionType)
	GetExtensionTypes() ExtensionTypes
	GetDefinitions() ExtensionDefinitions
}

type _ExtensionRegistry struct {
	lock       sync.Mutex
	extensions ExtensionTypes
}

func NewExtensionRegistry() ExtensionRegistry {
	return &_ExtensionRegistry{extensions: ExtensionTypes{}}
}

func (this *_ExtensionRegistry) RegisterExtension(e ExtensionType) error {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.extensions[e.Name()] != nil {
		return fmt.Errorf("extension with name %q already registered", e.Name())
	}
	this.extensions[e.Name()] = e
	return nil
}

func (this *_ExtensionRegistry) MustRegisterExtension(e ExtensionType) {
	if err := this.RegisterExtension(e); err != nil {
		panic(err)
	}
}

func (this *_ExtensionRegistry) GetExtensionTypes() ExtensionTypes {
	this.lock.Lock()
	defer this.lock.Unlock()

	ext := ExtensionTypes{}
	for n, t := range this.extensions {
		ext[n] = t
	}
	return ext
}

func (this *_ExtensionRegistry) GetDefinitions() ExtensionDefinitions {
	this.lock.Lock()
	defer this.lock.Unlock()

	ext := ExtensionDefinitions{}
	for n, e := range this.extensions {
		ext[n] = e.Definition()
	}
	return ext
}

var extensions = NewExtensionRegistry()

func DefaultRegistry() ExtensionRegistry {
	return extensions
}

func RegisterExtension(e ExtensionType) {
	extensions.RegisterExtension(e)
}

////////////////////////////////////////////////////////////////////////////////

type ControllerManager interface {
	GetName() string
	GetNamespace() string

	GetConfig() *areacfg.Config
	NewContext(key, value string) logger.LogContext
	GetContext() context.Context

	GetCluster(name string) cluster.Interface
	GetClusters() cluster.Clusters
	ClusterDefinitions() cluster.Definitions

	GetExtension(name string) Extension
}

type Environment interface {
	logger.LogContext
	Name() string
	Namespace() string
	GetContext() context.Context
	GetCluster(name string) cluster.Interface
	GetClusters() cluster.Clusters
	ClusterDefinitions() cluster.Definitions
}

type environment struct {
	logger.LogContext
	name    string
	context context.Context
	manager ControllerManager
}

func NewDefaultEnvironment(ctx context.Context, name string, manager ControllerManager) Environment {
	if ctx == nil {
		ctx = manager.GetContext()
	}
	logctx := manager.NewContext("extension", name)
	return &environment{
		LogContext: logctx,
		name:       name,
		context:    logger.Set(ctx, logctx),
		manager:    manager,
	}
}

func (this *environment) Name() string {
	return this.name
}

func (this *environment) Namespace() string {
	return this.manager.GetNamespace()
}

func (this *environment) GetContext() context.Context {
	return this.context
}

func (this *environment) GetCluster(name string) cluster.Interface {
	return this.manager.GetCluster(name)
}

func (this *environment) GetClusters() cluster.Clusters {
	return this.manager.GetClusters()
}

func (this *environment) ClusterDefinitions() cluster.Definitions {
	return this.manager.ClusterDefinitions()
}

////////////////////////////////////////////////////////////////////////////////

type ElementBase interface {
	logger.LogContext

	GetType() string
	GetName() string

	GetContext() context.Context

	GetOption(name string) (*config.ArbitraryOption, error)
	GetBoolOption(name string) (bool, error)
	GetStringOption(name string) (string, error)
	GetStringArrayOption(name string) ([]string, error)
	GetIntOption(name string) (int, error)
	GetDurationOption(name string) (time.Duration, error)
}

type elementBase struct {
	logger.LogContext
	name     string
	typeName string
	context  context.Context
	options  config.Options
}

func NewElementBase(ctx context.Context, valueType ctxutil.ValueKey, element interface{}, name string, set config.Options) ElementBase {
	ctx = valueType.WithValue(ctx, name)
	ctx, logctx := logger.WithLogger(ctx, valueType.Name(), name)
	return &elementBase{
		LogContext: logctx,
		context:    ctx,
		name:       name,
		typeName:   valueType.Name(),
		options:    set,
	}
}

func (this *elementBase) GetName() string {
	return this.name
}

func (this *elementBase) GetType() string {
	return this.typeName
}

func (this *elementBase) GetContext() context.Context {
	return this.context
}

func (this *elementBase) GetOption(name string) (*config.ArbitraryOption, error) {
	opt := this.options.GetOption(name)
	if opt == nil {
		return nil, fmt.Errorf("unknown option %q for %s %q", name, this.GetType(), this.GetName())
	}
	return opt, nil
}

func (this *elementBase) GetBoolOption(name string) (bool, error) {
	opt, err := this.GetOption(name)
	if err != nil {
		return false, err
	}
	return opt.BoolValue(), nil
}

func (this *elementBase) GetStringOption(name string) (string, error) {
	opt, err := this.GetOption(name)
	if err != nil {
		return "", err
	}
	return opt.StringValue(), nil
}

func (this *elementBase) GetStringArrayOption(name string) ([]string, error) {
	opt, err := this.GetOption(name)
	if err != nil {
		return []string{}, err
	}
	return opt.StringArray(), nil
}

func (this *elementBase) GetIntOption(name string) (int, error) {
	opt, err := this.GetOption(name)
	if err != nil {
		return 0, err
	}
	return opt.IntValue(), nil
}

func (this *elementBase) GetDurationOption(name string) (time.Duration, error) {
	opt, err := this.GetOption(name)
	if err != nil {
		return 0, err
	}
	return opt.DurationValue(), nil
}

////////////////////////////////////////////////////////////////////////////////

type OptionDefinition interface {
	GetName() string
	Type() config.OptionType
	Default() interface{}
	Description() string
}

type OptionDefinitions map[string]OptionDefinition

///////////////////////////////////////////////////////////////////////////////

type DefaultOptionDefinition struct {
	name         string
	gotype       config.OptionType
	defaultValue interface{}
	desc         string
}

func NewOptionDefinition(name string, gotype config.OptionType, def interface{}, desc string) OptionDefinition {
	return &DefaultOptionDefinition{name, gotype, def, desc}
}

func (this *DefaultOptionDefinition) GetName() string {
	return this.name
}

func (this *DefaultOptionDefinition) Type() config.OptionType {
	return this.gotype
}

func (this *DefaultOptionDefinition) Default() interface{} {
	return this.defaultValue
}

func (this *DefaultOptionDefinition) Description() string {
	return this.desc
}

var _ OptionDefinition = &DefaultOptionDefinition{}
