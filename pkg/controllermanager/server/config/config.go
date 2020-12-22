/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"github.com/gardener/controller-manager-library/pkg/config"
	areacfg "github.com/gardener/controller-manager-library/pkg/controllermanager/config"
)

const OPTION_SOURCE = "server"

type Config struct {
	Servers string

	config.OptionSet
}

var _ config.OptionSource = (*Config)(nil)

func NewConfig() *Config {
	cfg := &Config{
		OptionSet: config.NewDefaultOptionSet(OPTION_SOURCE, OPTION_SOURCE),
	}
	cfg.AddStringOption(&cfg.Servers, "servers", "w", "all", "comma separated list of servers to start (<name>,<group>,all)")
	return cfg
}

func (this *Config) AddOptionsToSet(set config.OptionSet) {
	this.OptionSet.AddOptionsToSet(set)
}

func GetConfig(cfg *areacfg.Config) *Config {
	return cfg.GetSource(OPTION_SOURCE).(*Config)
}
