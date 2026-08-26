/*
 * SPDX-FileCopyrightText: Copyright Contributors to the Gardener project
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package configmain

import (
	"github.com/gardener/controller-manager-library/pkg/config"
)

const OPTION_SOURCE = "main"

type Config struct {
	evaluated bool
	*config.DefaultOptionSet
}

func NewConfig() *Config {
	cfg := &Config{}
	cfg.DefaultOptionSet = config.NewDefaultOptionSet(OPTION_SOURCE, "")
	addExtensions(cfg)
	return cfg
}

func (this *Config) Evaluate() error {
	if !this.evaluated {
		this.evaluated = true
		return this.DefaultOptionSet.Evaluate()
	}
	return nil
}
