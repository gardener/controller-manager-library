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

package extension

import (
	. "github.com/onsi/ginkgo"
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
		Expect(err).To(BeNil())
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
		Expect(err).To(BeNil())
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
