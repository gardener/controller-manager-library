/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package cluster

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/gardener/controller-manager-library/pkg/resources"
	"github.com/gardener/controller-manager-library/pkg/resources/schemes"
	"github.com/gardener/controller-manager-library/pkg/utils"
)

///////////////////////////////////////////////////////////////////////////////
// cluster definitions
///////////////////////////////////////////////////////////////////////////////

type Registrations map[string]Definition

type Registerable interface {
	Definition() Definition
}

type RegistrationInterface interface {
	RegisterCluster(Registerable) error
	MustRegisterCluster(Registerable) RegistrationInterface
}

type Registry interface {
	RegistrationInterface
	GetDefinitions() Definitions
}

type _Registry struct {
	*_Definitions
}

func NewRegistry(src resources.SchemeSource) Registry {
	if src == nil {
		src = resources.DefaultSchemeSource()
	}
	registry := &_Registry{_Definitions: &_Definitions{definitions: Registrations{}, schemeSource: src}}
	Configure(DEFAULT, "kubeconfig", "default cluster access").MustRegisterAt(registry)
	return registry
}

func NewRegistryForScheme(scheme *runtime.Scheme) Registry {
	var src resources.SchemeSource
	if scheme != nil {
		src = schemes.ExplicitSchemeSource(scheme, "explicit")
	}
	return NewRegistry(src)
}

////////////////////////////////////////////////////////////////////////////////

var _ Registry = &_Registry{}

func (this *_Registry) RegisterCluster(reg Registerable) error {
	def := reg.Definition()
	if def == nil {
		return fmt.Errorf("no definition found")
	}
	this.lock.Lock()
	defer this.lock.Unlock()

	if old := this.definitions[def.Name()]; old != nil {
		msg := fmt.Sprintf("cluster request for %q", def.Name())
		new := copy(old)
		err := utils.FillStringValue(msg, &new.configOptionName, def.ConfigOptionName())
		if err != nil {
			return err
		}
		err = utils.FillStringValue(msg, &new.description, def.Description())
		if err != nil {
			return err
		}
		def = new
	}
	this.definitions[def.Name()] = def
	return nil
}

func (this *_Registry) MustRegisterCluster(reg Registerable) RegistrationInterface {
	err := this.RegisterCluster(reg)
	if err != nil {
		panic(err)
	}
	return this
}

func (this *_Registry) GetDefinitions() Definitions {
	defs := Registrations{}
	for k, v := range this.definitions {
		defs[k] = v
	}
	return &_Definitions{definitions: defs}
}
