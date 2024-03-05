/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 *
 */

package utils

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type Struct struct{}

// nolint:staticcheck
var _ = Describe("Types", func() {

	Context("Nil interface", func() {

		It("should handle nil", func() {
			var i interface{}

			Expect(i).To(BeNil())

			Expect(IsNil(i)).To(BeTrue())
		})

		It("should handle pointers", func() {
			var i interface{}

			var p *Struct
			i = p
			if i == nil {
				panic("NIL")
			}
			Expect(IsNil(i)).To(BeTrue())
		})

		It("should handle slice", func() {
			var i interface{}

			var p []Struct
			i = p
			if i == nil {
				panic("NIL")
			}

			Expect(IsNil(i)).To(BeTrue())
		})

		It("should handle function", func() {
			var i interface{}

			var p func(bool) bool
			i = p
			if i == nil {
				panic("NIL")
			}

			Expect(IsNil(i)).To(BeTrue())
		})

		It("should handle chan", func() {
			var i interface{}

			var p chan bool
			i = p
			if i == nil {
				panic("NIL")
			}

			Expect(IsNil(i)).To(BeTrue())
		})

		It("should handle map", func() {
			var i interface{}

			var p map[string]Struct
			i = p
			if i == nil {
				panic("NIL")
			}
			Expect(IsNil(i)).To(BeTrue())
		})
	})
})
