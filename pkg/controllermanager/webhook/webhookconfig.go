/*
 * Copyright 2020 SAP SE or an SAP affiliate company. All rights reserved.
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

	"github.com/gardener/controller-manager-library/pkg/config"
	areacfg "github.com/gardener/controller-manager-library/pkg/controllermanager/webhook/config"
)

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
	for name, def := range this.definitions {
		wcfg := NewWebhookConfig(name)
		cfg.AddSource(name, wcfg)

		for oname, o := range def.ConfigOptions() {
			wcfg.AddOption(o.Type(), nil, oname, "", o.Default(), o.Description())
		}
	}
}
