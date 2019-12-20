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
	"github.com/gardener/controller-manager-library/pkg/controllermanager/config"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/utils"
	"sync"
)

type ExtensionType interface {
	Name() string
	Definition() ExtensionDefinition
}

type ExtensionDefinition interface {
	Name() string
	Names() utils.StringSet
	Size() int
	ExtendConfig(*config.Config)
	CreateExtension(logctx logger.LogContext, cm *ControllerManager) (Extension, error)
}

type Extension interface {
	Name() string
	RequiredClusters() (utils.StringSet, error)
	Start(ctx context.Context) error
}

type ExtensionRegistry interface {
	RegisterExtension(e ExtensionType)
	GetExtensionTypes() []ExtensionType
	GetDefinitions() []ExtensionDefinition
}

type _ExtensionRegistry struct {
	lock       sync.Mutex
	extensions []ExtensionType
}

func NewExtensionRegistry() ExtensionRegistry {
	return &_ExtensionRegistry{}
}

func (this *_ExtensionRegistry) RegisterExtension(e ExtensionType) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.extensions = append(this.extensions, e)
}

func (this *_ExtensionRegistry) GetExtensionTypes() []ExtensionType {
	this.lock.Lock()
	defer this.lock.Unlock()

	ext := make([]ExtensionType, len(this.extensions))
	copy(ext, this.extensions)
	return ext
}

func (this *_ExtensionRegistry) GetDefinitions() []ExtensionDefinition {
	this.lock.Lock()
	defer this.lock.Unlock()

	ext := make([]ExtensionDefinition, len(this.extensions))
	for i, e := range this.extensions {
		ext[i] = e.Definition()
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
