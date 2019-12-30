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
	"strings"

	areacfg "github.com/gardener/controller-manager-library/pkg/controllermanager/webhooks/config"
	"github.com/gardener/controller-manager-library/pkg/utils"
)

type Definitions interface {
	Get(name string) Definition
	Size() int
	Names() utils.StringSet
	Registrations(names ...string) (Registrations, error)
	ExtendConfig(cfg *areacfg.Config)

	GetActiveWebhooks(spec string) (utils.StringSet, error)
}

func (this *_Definitions) Size() int {
	return len(this.definitions)
}

func (this *_Definitions) Names() utils.StringSet {
	set := utils.StringSet{}
	for n := range this.definitions {
		set.Add(n)
	}
	return set
}

func (this *_Definitions) Registrations(names ...string) (Registrations, error) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	var r = Registrations{}

	if len(names) == 0 {
		r = this.definitions.Copy()
	} else {
		for _, name := range names {
			def := this.definitions[name]
			if def == nil {
				return nil, fmt.Errorf("webhook %q not found", name)
			}
			r[name] = def
		}
	}
	return r, nil
}

func (this *_Definitions) ExtendConfig(cfg *areacfg.Config) {

}

func (this *_Definitions) GetActiveWebhooks(active string) (utils.StringSet, error) {
	result := utils.StringSet{}
	for _, w := range strings.Split(active, ",") {
		w = strings.TrimSpace(w)
		if w == "all" {
			result.AddSet(this.Names())
		} else {
			if this.Get(w) == nil {
				return nil, fmt.Errorf("unknown webhook %q", w)
			}
			result.Add(w)
		}
	}
	return result, nil
}
