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

package utils

type Struct struct{}

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
