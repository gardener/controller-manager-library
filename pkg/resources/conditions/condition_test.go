/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 *
 */

package conditions_test

import (
	"fmt"
	"time"

	"github.com/gardener/controller-manager-library/pkg/resources/conditions"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
	Expect(my.Status.Conditions).To(HaveLen(length))
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
			Expect(type1.SetStatus(my, "True")).NotTo(HaveOccurred())
			Expect(type2.SetStatus(my, "False")).NotTo(HaveOccurred())
			checkCondition("create for status", my, now, 2,
				&MyCondition{Type: type1.Name(), Status: "True"},
				&MyCondition{Type: type2.Name(), Status: "False"},
			)

			Expect(type1.SetStatus(my, "False")).NotTo(HaveOccurred())
			Expect(type2.SetStatus(my, "True")).NotTo(HaveOccurred())
			checkCondition("set status", my, now, 2,
				&MyCondition{Type: type1.Name(), Status: "False"},
				&MyCondition{Type: type2.Name(), Status: "True"},
			)
		})

		It("add and updates reason", func() {
			my := &My{}

			now := time.Now()
			Expect(type1.SetReason(my, "True")).NotTo(HaveOccurred())
			Expect(type2.SetReason(my, "False")).NotTo(HaveOccurred())
			checkCondition("create for reason", my, now, 2,
				&MyCondition{Type: type1.Name(), Reason: "True"},
				&MyCondition{Type: type2.Name(), Reason: "False"},
			)

			Expect(type1.SetReason(my, "False")).NotTo(HaveOccurred())
			Expect(type2.SetReason(my, "True")).NotTo(HaveOccurred())
			checkCondition("set reason", my, now, 2,
				&MyCondition{Type: type1.Name(), Reason: "False"},
				&MyCondition{Type: type2.Name(), Reason: "True"},
			)
		})
		It("add and updates message", func() {
			my := &My{}

			now := time.Now()
			Expect(type1.SetMessage(my, "True")).NotTo(HaveOccurred())
			Expect(type2.SetMessage(my, "False")).NotTo(HaveOccurred())
			checkCondition("create for message", my, now, 2,
				&MyCondition{Type: type1.Name(), Message: "True"},
				&MyCondition{Type: type2.Name(), Message: "False"},
			)

			Expect(type1.SetMessage(my, "False")).NotTo(HaveOccurred())
			Expect(type2.SetMessage(my, "True")).NotTo(HaveOccurred())
			checkCondition("set message", my, now, 2,
				&MyCondition{Type: type1.Name(), Message: "False"},
				&MyCondition{Type: type2.Name(), Message: "True"},
			)
		})

		It("delete type", func() {
			my := &My{}

			now := time.Now()
			Expect(type1.SetStatus(my, "True")).NotTo(HaveOccurred())
			Expect(type2.SetStatus(my, "False")).NotTo(HaveOccurred())
			Expect(type3.SetStatus(my, "Other")).NotTo(HaveOccurred())
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
			Expect(cond1.SetStatus("True")).NotTo(HaveOccurred())

			cond2 := conds.Get(type2.Name())
			Expect(cond2, err).NotTo(BeNil())
			Expect(cond2.SetStatus("False")).NotTo(HaveOccurred())

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

			Expect(cond1.SetStatus("True")).NotTo(HaveOccurred())
			Expect(cond2.SetStatus("False")).NotTo(HaveOccurred())

			conds.ResetModified()
			Expect(conds.IsModified()).To(BeFalse())
			Expect(cond2.SetReason("reason")).NotTo(HaveOccurred())
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
			Expect(cond.SetStatus("A")).NotTo(HaveOccurred())

			ts := my.Status.Conditions[0].TransitionTime
			time.Sleep(10 * time.Nanosecond)

			Expect(cond.SetStatus("B")).NotTo(HaveOccurred())
			Expect(my.Status.Conditions[0].TransitionTime.Nanosecond()).To(BeNumerically(">", ts.Nanosecond()))
		})

		It("no update on no state change", func() {
			my := &My{}
			conds, _ := layout.For(my)

			cond := conds.Get(type1.Name())
			Expect(cond.SetStatus("A")).NotTo(HaveOccurred())

			ts := my.Status.Conditions[0].TransitionTime

			Expect(cond.SetStatus("A")).NotTo(HaveOccurred())
			Expect(my.Status.Conditions[0].TransitionTime).To(Equal(ts))
		})

		It("no update on other change", func() {
			my := &My{}
			conds, _ := layout.For(my)

			cond := conds.Get(type1.Name())
			Expect(cond.SetStatus("A")).NotTo(HaveOccurred())

			ts := my.Status.Conditions[0].TransitionTime

			Expect(cond.SetReason("x")).NotTo(HaveOccurred())
			Expect(my.Status.Conditions[0].TransitionTime).To(Equal(ts))
		})
	})

})
