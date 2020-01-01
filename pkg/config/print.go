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

package config

import "fmt"

type Writer func(msgfmt string, args ...interface{})

func PrintfWriter(msgfmt string, args ...interface{}) {
	fmt.Printf(msgfmt+"\n", args...)
}

func Print(log Writer, gap string, grp OptionSourceGroup) {
	log("%s* %s", gap, grp.Name())
	if set, ok := grp.(OptionSet); ok {
		set.VisitOptions(func(o *ArbitraryOption) bool {
			log("%s  %s: %t: %v (%s)", gap, o.Name, o.Changed(), o.Value(), o.Description)
			return true
		})
	}
	grp.VisitSources(func(key string, t OptionSource) bool {
		if sub, ok := t.(OptionSourceGroup); ok {
			Print(log, gap+"  ", sub)
		}
		return true
	})
}
