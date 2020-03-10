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
	"fmt"
	"sync"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/webhook/groups"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/utils"
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
	Register(reg Registerable, group ...string) error
	MustRegister(reg Registerable, group ...string) RegistrationInterface
}

type Registry interface {
	RegistrationInterface
	GetDefinitions() Definitions
}

type _Definitions struct {
	lock        sync.RWMutex
	definitions Registrations
	groups      groups.Definitions
}

type _Registry struct {
	*_Definitions
	groups groups.Registry
}

var _ Definition = &_Definition{}
var _ Definitions = &_Definitions{}

func NewRegistry() Registry {
	return newRegistry(groups.NewRegistry())
}

func newRegistry(groups groups.Registry) Registry {
	return &_Registry{_Definitions: &_Definitions{definitions: Registrations{}}, groups: groups}
}

func DefaultDefinitions() Definitions {
	return registry.GetDefinitions()
}

func DefaultRegistry() Registry {
	return registry
}

////////////////////////////////////////////////////////////////////////////////

var _ Registry = &_Registry{}

func (this *_Registry) Register(reg Registerable, grps ...string) error {
	def := reg.Definition()
	if def == nil {
		return fmt.Errorf("no Definition found")
	}
	this.lock.Lock()
	defer this.lock.Unlock()

	if d, ok := this.definitions[def.Name()]; ok && d != def {
		return fmt.Errorf("multiple registrations of webhook %q", def.Name())
	}
	logger.Infof("Registering webhook %s", def.Name())

	if len(grps) == 0 {
		grps = []string{groups.DEFAULT, string(def.Kind())}
	} else {
		grps = append(grps, string(def.Kind()))
	}
	for _, g := range grps {
		err := this.addToGroup(def, g)
		if err != nil {
			return err
		}
	}
	this.definitions[def.Name()] = def
	return nil
}

func (this *_Registry) MustRegister(reg Registerable, group ...string) RegistrationInterface {
	err := this.Register(reg, group...)
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
		groups:      this.groups.GetDefinitions(),
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

var registry = newRegistry(groups.DefaultRegistry())

func (this *_Registry) addToGroup(def Definition, name string) error {
	grp, err := this.groups.RegisterGroup(name)
	if err != nil {
		return err
	}
	if def.ActivateExplicitly() {
		grp.ActivateExplicitly(def.Name())
	}
	return grp.Members(def.Name())
}

///////////////////////////////////////////////////////////////////////////////

func Register(reg Registerable, group ...string) error {
	return registry.Register(reg, group...)
}

func MustRegister(reg Registerable, group ...string) RegistrationInterface {
	return registry.MustRegister(reg, group...)
}
