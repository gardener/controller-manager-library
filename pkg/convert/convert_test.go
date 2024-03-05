/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package convert

import (
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type Struct struct {
	value string
}

func (this *Struct) String() string {
	return this.value
}

type StringList []string
type StringLists []StringList

var _ = Describe("Conversions", func() {

	Context("explicit conversions", func() {
		type x int

		type f float64

		It("converts string", func() {
			v := "value"

			r := String(&v)
			Expect(r).To(Equal("value"))
		})

		It("converts bool", func() {
			v := true

			r := Bool(&v)
			Expect(r).To(BeTrue())
		})

		It("converts int", func() {
			v := 64

			r := Int(&v)
			Expect(r).To(Equal(64))
		})

		It("converts int32", func() {
			v := int32(64)

			r := Int32(&v)
			Expect(r).To(Equal(int32(64)))
		})
		It("converts int to int32", func() {
			v := 64

			r := Int32(&v)
			Expect(r).To(Equal(int32(64)))
		})

		It("converts int64", func() {
			v := int64(64)

			r := Int64(&v)
			Expect(r).To(Equal(int64(64)))
		})
		It("converts int to int64", func() {
			v := 64

			r := Int64(&v)
			Expect(r).To(Equal(int64(64)))
		})

		It("converts x to int64", func() {
			v := x(64)

			r := Int64(&v)
			Expect(r).To(Equal(int64(64)))
		})

		It("converts float64", func() {
			v := float64(3.14)

			r := Float64(&v)
			Expect(r).To(Equal(float64(3.14)))
		})
		It("converts int to float64", func() {
			v := 64

			r := Float64(&v)
			Expect(r).To(Equal(float64(64)))
		})

		It("converts f to float32", func() {
			v := f(3.14)

			r := Float32(&v)
			Expect(r).To(Equal(float32(3.14)))
		})
	})

	////////////////////////////////////////////////////////////////////////////

	Context("string", func() {
		Context("direct string", func() {
			It("converts string", func() {
				v := "value"

				r, err := ConvertTo(v, StringType())
				Expect(err).NotTo(HaveOccurred())
				Expect(r).To(Equal("value"))
			})

			It("converts string pointer", func() {
				v := "value"

				r, err := ConvertTo(&v, StringType())
				Expect(err).NotTo(HaveOccurred())
				Expect(r).To(Equal("value"))
			})

			It("converts string pointer pointer", func() {
				v := "value"
				p := &v
				r, err := ConvertTo(&p, StringType())
				Expect(err).NotTo(HaveOccurred())
				Expect(r).To(Equal("value"))
			})
		})

		Context("retyped string", func() {
			type x string

			It("converts string", func() {
				v := x("value")

				r, err := ConvertTo(v, StringType())
				Expect(err).NotTo(HaveOccurred())
				Expect(r).To(Equal("value"))
			})

			It("converts string pointer", func() {
				v := x("value")

				r, err := ConvertTo(&v, StringType())
				Expect(err).NotTo(HaveOccurred())
				Expect(r).To(Equal("value"))
			})

			It("converts string pointer pointer", func() {
				v := x("value")
				p := &v
				r, err := ConvertTo(&p, StringType())
				Expect(err).NotTo(HaveOccurred())
				Expect(r).To(Equal("value"))
			})
		})

	})

	////////////////////////////////////////////////////////////////////////////

	Context("struct", func() {
		var targetType = reflect.TypeOf(Struct{})

		Context("direct struct", func() {
			It("converts struct", func() {
				v := Struct{"value"}

				r, err := ConvertTo(v, targetType)
				Expect(err).NotTo(HaveOccurred())
				Expect(r).NotTo(BeNil())
				Expect(r).To(Equal(v))
			})

			It("converts struct pointer", func() {
				v := Struct{"value"}

				r, err := ConvertTo(&v, targetType)
				Expect(err).NotTo(HaveOccurred())
				Expect(r).NotTo(BeNil())
				Expect(r).To(Equal(v))
			})

			It("converts struct pointer pointer", func() {
				v := Struct{"value"}
				p := &v
				r, err := ConvertTo(&p, targetType)
				Expect(err).NotTo(HaveOccurred())
				Expect(r).NotTo(BeNil())
				Expect(r).To(Equal(v))
			})
		})

		Context("retyped struct", func() {
			type x Struct

			It("converts struct", func() {
				v := x(Struct{"value"})

				r, err := ConvertTo(v, targetType)
				Expect(err).NotTo(HaveOccurred())
				Expect(r).NotTo(BeNil())
				Expect(r).To(Equal(Struct(v)))
			})

			It("converts struct pointer", func() {
				v := x(Struct{"value"})

				r, err := ConvertTo(&v, targetType)
				Expect(err).NotTo(HaveOccurred())
				Expect(r).NotTo(BeNil())
				Expect(r).To(Equal(Struct(v)))
			})

			It("converts struct pointer pointer", func() {
				v := x(Struct{"value"})
				p := &v
				r, err := ConvertTo(&p, targetType)
				Expect(err).NotTo(HaveOccurred())
				Expect(r).NotTo(BeNil())
				Expect(r).To(Equal(Struct(v)))
			})
		})

	})

	////////////////////////////////////////////////////////////////////////////

	Context("slices", func() {
		type Slice []string

		Context("direct slice", func() {
			It("converts slice", func() {
				v := Slice{"value"}

				r, err := ConvertTo(v, Slice{})
				Expect(err).NotTo(HaveOccurred())
				Expect(r).To(Equal(Slice{"value"}))
			})
			It("converts retyped slice", func() {
				type XSlice Slice
				v := XSlice{"value"}

				r, err := ConvertTo(v, Slice{})
				Expect(err).NotTo(HaveOccurred())
				Expect(r).To(Equal(Slice{"value"}))
			})
		})

		Context("retyped slice elements", func() {
			type x string
			type XSlice []x

			It("converts slice", func() {
				v := XSlice{"value"}

				r, err := ConvertTo(v, Slice{})
				Expect(err).NotTo(HaveOccurred())
				Expect(r).To(Equal(Slice{"value"}))
			})
		})
	})
	////////////////////////////////////////////////////////////////////////////

	Context("maps", func() {
		type Map map[string]string

		Context("direct map", func() {
			It("converts map", func() {
				v := Map{"key": "value"}

				r, err := ConvertTo(v, Map{})
				Expect(err).NotTo(HaveOccurred())
				Expect(r).To(Equal(Map{"key": "value"}))
			})
			It("converts retyped map", func() {
				type XMap Map
				v := XMap{"key": "value"}

				r, err := ConvertTo(v, Map{})
				Expect(err).NotTo(HaveOccurred())
				Expect(r).To(Equal(Map{"key": "value"}))
			})
		})

		Context("retyped map elements", func() {
			type x string
			type XMap map[string]x

			It("converts map", func() {
				v := XMap{"key": "value"}

				r, err := ConvertTo(v, Map{})
				Expect(err).NotTo(HaveOccurred())
				Expect(r).To(Equal(Map{"key": "value"}))
			})
		})
	})

	Context("converts string listss", func() {
		lists := [][]string{
			{
				"a", "b",
			},
			{
				"c", "d",
			},
		}

		It("converts map", func() {
			v := StringLists{
				StringList{
					"a", "b",
				},
				StringList{
					"c", "d",
				},
			}

			r, err := ConvertTo(lists, StringLists{})
			Expect(err).NotTo(HaveOccurred())
			Expect(r).To(Equal(v))
		})
	})
})
