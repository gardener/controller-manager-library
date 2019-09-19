/*
 * Copyright 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package cond

import (
	"fmt"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller/reconcile/conditions"
	"time"
)

type My struct {
	Status MyStatus
}

type MyStatus struct {
	Conditions []MyCondition
}

type MyCondition struct {
	Type           string
	Message        string
	Status         string
	TransitionTime time.Time
}

func CondMain() {
	cond := conditions.NewCondition("test")
	my := &My{}

	my.Status.Conditions = append(my.Status.Conditions, MyCondition{Type: "test", Status: "done"})

	c := cond.GetInterface(my)
	if c == nil {
		fmt.Println("condition test not found")
	} else {
		fmt.Printf("Status: %s\n", c.(*MyCondition).Status)
	}

	my2 := &My{}

	cond.AssureInterface(my2).(*MyCondition).Message = "It works"

	cond.SetStatus(my2, "done")

	fmt.Printf("%#v\n", my2)

	my2 = &My{}

	cd := cond.GetCondition(my2)
	cd.AssureInterface().(*MyCondition).Message = "It works"

	cd.SetStatus("done")

	fmt.Printf("Message: %s\n", cd.GetMessage())
	fmt.Printf("Transition: %s\n", cd.GetTransitionTime())
	t:=cd.GetLastUpdateTime()
	if t.IsZero() {
		fmt.Printf("Update: not set\n")
	} else {
		fmt.Printf("Update: %s\n", t)
	}
	fmt.Printf("%#v\n", my2)
}
