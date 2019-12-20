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
	"fmt"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/utils"
	"sync"
)

///////////////////////////////////////////////////////////////////////////////
// controller Registrations
///////////////////////////////////////////////////////////////////////////////

type Registrations map[string]Definition

func (this Registrations) Copy() Registrations {
	r := Registrations{}
	for n, def := range this {
		r[n] = def
	}
	return r
}

func (this Registrations) Names() utils.StringSet {
	r := utils.StringSet{}
	for n := range this {
		r.Add(n)
	}
	return r
}

type Registerable interface {
	Definition() Definition
}

type RegistrationInterface interface {
	Register(Registerable) error
	MustRegister(Registerable) RegistrationInterface
}

type Registry interface {
	RegistrationInterface
	GetDefinitions() Definitions
}

type _Definitions struct {
	lock        sync.RWMutex
	definitions Registrations
}

type _Registry struct {
	*_Definitions
}

var _ Definition = &_Definition{}
var _ Definitions = &_Definitions{}

func NewRegistry() Registry {
	return &_Registry{_Definitions: &_Definitions{definitions: Registrations{}}}
}

func DefaultDefinitions() Definitions {
	return registry.GetDefinitions()
}

func DefaultRegistry() Registry {
	return registry
}

////////////////////////////////////////////////////////////////////////////////

var _ Registry = &_Registry{}

func (this *_Registry) Register(reg Registerable) error {
	def := reg.Definition()
	if def == nil {
		return fmt.Errorf("no Definition found")
	}
	this.lock.Lock()
	defer this.lock.Unlock()

	if d, ok := this.definitions[def.GetName()]; ok && d != def {
		return fmt.Errorf("multiple registrations of webhook %q", def.GetName())
	}
	logger.Infof("Registering webhook %s", def.GetName())

	this.definitions[def.GetName()] = def
	return nil
}

func (this *_Registry) MustRegister(reg Registerable) RegistrationInterface {
	err := this.Register(reg)
	if err != nil {
		panic(err)
	}
	return this
}

////////////////////////////////////////////////////////////////////////////////

func (this *_Registry) GetDefinitions() Definitions {
	defs := Registrations{}
	for k, v := range this.definitions {
		defs[k] = v
	}
	return &_Definitions{
		definitions: defs,
	}
}

func (this *_Definitions) Get(name string) Definition {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.definitions[name]
}

func (this *_Definition) Definition() Definition {
	return this
}

///////////////////////////////////////////////////////////////////////////////

var registry = NewRegistry()

func Register(reg Registerable) error {
	return registry.Register(reg)
}

func MustRegister(reg Registerable) RegistrationInterface {
	return registry.MustRegister(reg)
}
