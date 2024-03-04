/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package extension

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type Elem struct {
	name   string
	after  []string
	before []string
}

func (this *Elem) Name() string {
	return this.name
}

func (this *Elem) After() []string {
	return this.after
}

func (this *Elem) Before() []string {
	return this.before
}

func (this *Elem) String() string {
	return this.name
}

var _ = Describe("Order", func() {
	A := &Elem{name: "A", after: []string{"B", "C"}}
	B := &Elem{name: "B", after: []string{"D", "C"}}
	C := &Elem{name: "C", after: []string{"D"}}
	D := &Elem{name: "D"}
	D1 := &Elem{name: "D", after: []string{"B"}}

	_ = D1
	It("sort array after", func() {
		o, _, err := Order([]*Elem{
			A, B, C, D,
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(o).To(Equal([]string{"D", "C", "B", "A"}))
	})
	It("sort map after", func() {
		m := map[string]*Elem{
			"A": A,
			"B": B,
			"C": C,
			"D": D,
		}
		o, _, err := Order(m)
		Expect(err).ToNot(HaveOccurred())
		Expect(o).To(Equal([]string{"D", "C", "B", "A"}))
	})

	It("detect cycle", func() {
		o, _, err := Order([]*Elem{
			A, B, C, D1,
		})
		Expect(o).To(BeNil())
		Expect(err.Error()).To(Equal("cycle detected: [B D B]"))
	})
})
