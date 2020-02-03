/*
 * Copyright 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 *
 */

package conditions_test

import (
	"fmt"
	"time"

	"github.com/gardener/controller-manager-library/pkg/resources/conditions"
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
	Reason         string
	Status         string
	TransitionTime time.Time
	LastUpdateTime time.Time
}

func checkCondition(msg string, my *My, now time.Time, length int, conds ...*MyCondition) {
	Expect(len(my.Status.Conditions)).To(Equal(length))
next:
	for _, cond := range conds {
		for _, c := range my.Status.Conditions {
			if c.Type == cond.Type {
				Expect(c.Status).To(Equal(cond.Status))
				Expect(c.Reason).To(Equal(cond.Reason))
				Expect(c.Message).To(Equal(cond.Message))
				Expect(c.LastUpdateTime.Nanosecond()).Should(BeNumerically(">", now.Nanosecond()))
				continue next
			}
		}
		Fail(fmt.Sprintf("%s: condition %q not found", msg, cond.Type), 1)
	}
}

var _ = Describe("Conditions", func() {

	var layout = conditions.NewConditionLayout()
	var type1 = conditions.NewConditionType("type1", layout)
	var type2 = conditions.NewConditionType("type2", layout)
	var type3 = conditions.NewConditionType("type3", layout)

	Context("Layout", func() {
		It("add and updates status", func() {
			my := &My{}

			now := time.Now()
			type1.SetStatus(my, "True")
			type2.SetStatus(my, "False")
			checkCondition("create for status", my, now, 2,
				&MyCondition{Type: type1.Name(), Status: "True"},
				&MyCondition{Type: type2.Name(), Status: "False"},
			)

			type1.SetStatus(my, "False")
			type2.SetStatus(my, "True")
			checkCondition("set status", my, now, 2,
				&MyCondition{Type: type1.Name(), Status: "False"},
				&MyCondition{Type: type2.Name(), Status: "True"},
			)
		})

		It("add and updates reason", func() {
			my := &My{}

			now := time.Now()
			type1.SetReason(my, "True")
			type2.SetReason(my, "False")
			checkCondition("create for reason", my, now, 2,
				&MyCondition{Type: type1.Name(), Reason: "True"},
				&MyCondition{Type: type2.Name(), Reason: "False"},
			)

			type1.SetReason(my, "False")
			type2.SetReason(my, "True")
			checkCondition("set reason", my, now, 2,
				&MyCondition{Type: type1.Name(), Reason: "False"},
				&MyCondition{Type: type2.Name(), Reason: "True"},
			)
		})
		It("add and updates message", func() {
			my := &My{}

			now := time.Now()
			type1.SetMessage(my, "True")
			type2.SetMessage(my, "False")
			checkCondition("create for message", my, now, 2,
				&MyCondition{Type: type1.Name(), Message: "True"},
				&MyCondition{Type: type2.Name(), Message: "False"},
			)

			type1.SetMessage(my, "False")
			type2.SetMessage(my, "True")
			checkCondition("set message", my, now, 2,
				&MyCondition{Type: type1.Name(), Message: "False"},
				&MyCondition{Type: type2.Name(), Message: "True"},
			)
		})

		It("delete type", func() {
			my := &My{}

			now := time.Now()
			type1.SetStatus(my, "True")
			type2.SetStatus(my, "False")
			type3.SetStatus(my, "Other")
			checkCondition("delete create", my, now, 3,
				&MyCondition{Type: type1.Name(), Status: "True"},
				&MyCondition{Type: type2.Name(), Status: "False"},
				&MyCondition{Type: type3.Name(), Status: "Other"},
			)

			type2.DeleteCondition(my)
			checkCondition("delete 2", my, now, 2,
				&MyCondition{Type: type1.Name(), Status: "True"},
				&MyCondition{Type: type3.Name(), Status: "Other"},
			)

			type1.DeleteCondition(my)
			checkCondition("delete 1", my, now, 1,
				&MyCondition{Type: type3.Name(), Status: "Other"},
			)
			type3.DeleteCondition(my)
			checkCondition("delete 3", my, now, 0)
		})
	})
	Context("Conditions", func() {
		It("creates conditions", func() {
			my := &My{}
			conds, err := layout.For(my)
			Expect(conds, err).NotTo(BeNil())
		})

		It("creates conditions", func() {
			now := time.Now()
			my := &My{}
			conds, err := layout.For(my)

			cond1 := conds.Get(type1.Name())
			Expect(cond1, err).NotTo(BeNil())
			cond1.SetStatus("True")

			cond2 := conds.Get(type2.Name())
			Expect(cond2, err).NotTo(BeNil())
			cond2.SetStatus("False")

			checkCondition("set status", my, now, 2,
				&MyCondition{Type: type1.Name(), Status: "True"},
				&MyCondition{Type: type2.Name(), Status: "False"},
			)
			Expect(conds.IsModified()).To(BeTrue())

			Expect(conds.Get(type1.Name())).To(BeIdenticalTo(cond1))
			Expect(conds.Get(type2.Name())).To(BeIdenticalTo(cond2))
		})
		It("updates conditions", func() {
			now := time.Now()
			my := &My{}
			conds, _ := layout.For(my)

			cond1 := conds.Get(type1.Name())
			cond2 := conds.Get(type2.Name())

			cond1.SetStatus("True")
			cond2.SetStatus("False")

			conds.ResetModified()
			Expect(conds.IsModified()).To(BeFalse())
			cond2.SetReason("reason")
			Expect(conds.IsModified()).To(BeTrue())

			checkCondition("set status", my, now, 2,
				&MyCondition{Type: type1.Name(), Status: "True"},
				&MyCondition{Type: type2.Name(), Status: "False", Reason: "reason"},
			)
		})

	})
	Context("Timestamps", func() {
		It("update on state change", func() {
			my := &My{}
			conds, _ := layout.For(my)

			cond := conds.Get(type1.Name())
			cond.SetStatus("A")

			ts := my.Status.Conditions[0].TransitionTime

			cond.SetStatus("B")
			Expect(my.Status.Conditions[0].TransitionTime.Nanosecond()).To(BeNumerically(">", ts.Nanosecond()))
		})

		It("no update on no state change", func() {
			my := &My{}
			conds, _ := layout.For(my)

			cond := conds.Get(type1.Name())
			cond.SetStatus("A")

			ts := my.Status.Conditions[0].TransitionTime

			cond.SetStatus("A")
			Expect(my.Status.Conditions[0].TransitionTime).To(Equal(ts))
		})

		It("no update on other change", func() {
			my := &My{}
			conds, _ := layout.For(my)

			cond := conds.Get(type1.Name())
			cond.SetStatus("A")

			ts := my.Status.Conditions[0].TransitionTime

			cond.SetReason("x")
			Expect(my.Status.Conditions[0].TransitionTime).To(Equal(ts))
		})
	})

})
