/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package test

import (
	"github.com/gardener/controller-manager-library/pkg/config"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/examples"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/module"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/module/handler"
)

func init() {
	module.Configure("test").
		OptionsByExample("options", &Config{}).
		RegisterHandler("testhandler", Create).
		MustRegister()
}

func Create(mod module.Interface) (handler.Interface, error) {
	config, err := mod.GetOptionSource("options")
	if err != nil {
		return nil, err
	}
	mod.Infof("found option message: %s", config.(*Config).message)

	return &Handler{
		module: mod,
		config: config.(*Config),
	}, nil
}

type Config struct {
	message string
}

func (this *Config) AddOptionsToSet(set config.OptionSet) {
	set.AddStringOption(&this.message, "message", "", "", "test message")
}

type Handler struct {
	module module.Interface
	config *Config
	shared *examples.SharedInfo
}

func (this *Handler) Setup() error {
	this.shared = examples.GetOrCreateShared(this.module.GetEnvironment())
	this.shared.Values["message"] = this.config.message
	return nil
}
