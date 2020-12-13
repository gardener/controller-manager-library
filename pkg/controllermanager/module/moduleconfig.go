/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package module

import (
	"fmt"

	"github.com/gardener/controller-manager-library/pkg/config"
	areacfg "github.com/gardener/controller-manager-library/pkg/controllermanager/module/config"
)

const MODULE_SET_PREFIX = "module."

type ModuleConfig struct {
	config.OptionSet
}

func NewModuleConfig(name string) *ModuleConfig {
	return &ModuleConfig{
		OptionSet: config.NewSharedOptionSet(name, name, func(desc string) string {
			return fmt.Sprintf("%s of module %s", desc, name)
		}),
	}
}

func (this *ModuleConfig) AddOptionsToSet(set config.OptionSet) {
	this.OptionSet.AddOptionsToSet(set)
}

func (this *ModuleConfig) Evaluate() error {
	return this.OptionSet.Evaluate()
}

func (this *_Definitions) ExtendConfig(cfg *areacfg.Config) {

	for name, def := range this.definitions {
		ccfg := NewModuleConfig(name)
		cfg.AddSource(name, ccfg)

		set := ccfg.OptionSet

		for oname, o := range def.ConfigOptions() {
			set.AddOption(o.Type(), nil, oname, "", o.Default(), o.Description())
		}
		for oname, o := range def.ConfigOptionSources() {
			if src := o.Create(); src != nil {
				set.AddSource(MODULE_SET_PREFIX+oname, src)
			}
		}
	}
}
