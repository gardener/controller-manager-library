/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package webhook

import (
	"fmt"

	"github.com/gardener/controller-manager-library/pkg/config"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/extension"
	areacfg "github.com/gardener/controller-manager-library/pkg/controllermanager/webhook/config"
	"github.com/gardener/controller-manager-library/pkg/utils"
)

const WEBHOOK_SET_PREFIX = "webhook."

type WebhookConfig struct {
	config.OptionSet
}

func NewWebhookConfig(name string) *WebhookConfig {
	return &WebhookConfig{
		OptionSet: config.NewSharedOptionSet(name, name, func(desc string) string {
			return fmt.Sprintf("%s of webhook %s", desc, name)
		}),
	}
}

func (this *_Definitions) ExtendConfig(cfg *areacfg.Config) {
	set := utils.StringSet{}
	for name, def := range this.definitions {
		wcfg := NewWebhookConfig(name)
		cfg.AddSource(name, wcfg)

		extension.AddElementConfigDefinitionToSet(def, WEBHOOK_SET_PREFIX, wcfg.OptionSet)
		kind := string(def.Kind())
		if !set.Contains(kind) {
			set.Add(kind)
			handler := GetRegistrationHandler(def.Kind())
			if handler != nil {
				os := handler.OptionSourceCreator()
				if os != nil {
					cfg.AddSource(kind, os())
				}
			}
		}
	}
}
