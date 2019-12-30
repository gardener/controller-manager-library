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

package controllermanager

import (
	"context"
	"fmt"
	"sync"

	areacfg "github.com/gardener/controller-manager-library/pkg/controllermanager/config"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/utils"
)

type ExtensionDefinitions map[string]ExtensionDefinition
type ExtensionTypes map[string]ExtensionType

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
	CreateExtension(logctx logger.LogContext, cm *ControllerManager) (Extension, error)
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
