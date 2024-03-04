/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package cond

import (
	"context"
	"fmt"
	"time"

	"github.com/gardener/controller-manager-library/pkg/fieldpath"
	"github.com/gardener/controller-manager-library/pkg/resources/conditions"
	"github.com/gardener/controller-manager-library/pkg/resources/plain"
	v1 "k8s.io/api/core/v1"
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
	stype := conditions.NewConditionLayout()
	ctype := conditions.NewConditionType("test", stype)
	my := &My{}

	my.Status.Conditions = append(my.Status.Conditions, MyCondition{Type: "test", Status: "done"})

	c := ctype.GetInterface(my)
	if c == nil {
		fmt.Println("condition test not found")
	} else {
		fmt.Printf("Status: %s\n", c.(*MyCondition).Status)
	}

	my2 := &My{}

	ctype.AssureInterface(my2).(*MyCondition).Message = "It works"

	_ = ctype.SetStatus(my2, "done")

	fmt.Printf("%#v\n", my2)

	my2 = &My{}

	cd := ctype.Get(my2)
	cd.AssureInterface().(*MyCondition).Message = "It works"

	_ = cd.SetStatus("done")

	fmt.Printf("Message: %s\n", cd.GetMessage())
	fmt.Printf("Transition: %s\n", cd.GetTransitionTime())
	t := cd.GetLastUpdateTime()
	if t.IsZero() {
		fmt.Printf("Update: not set\n")
	} else {
		fmt.Printf("Update: %s\n", t)
	}
	fmt.Printf("%t: %#v\n", cd.IsModified(), my2)

	cd = ctype.Get(my2)
	_ = cd.SetStatus("done")
	fmt.Printf("modified %t\n", cd.IsModified())

	f, err := fieldpath.NewField(&My{}, ".Status.Conditions[.Type=\"test\"].Bla")
	if err != nil {
		fmt.Printf("err: %s\n", err)
	} else {
		my := &My{}
		_ = f.Set(my, "it works")
		fmt.Printf("%#v\n", my)
	}

	podt := conditions.NewConditionLayout(conditions.TransitionTimeField("LastTransitionTime"))
	podc := conditions.NewConditionType("Test", podt)
	pod := &v1.Pod{}
	resc := plain.NewResourceContext(context.TODO(), nil).Resources()
	obj, err := resc.Wrap(pod)
	if err != nil {
		fmt.Printf("err: %s\n", err)
	}
	mod := plain.NewModificationState(obj)
	cd = mod.Condition(podc)
	_ = cd.SetMessage("test")
	err = cd.SetTransitionTime(time.Now())
	if err != nil {
		fmt.Printf("err: %s\n", err)
	}
	fmt.Printf("mod: %t: %#v\n", mod.IsModified(), pod)
}
