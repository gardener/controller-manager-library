/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package match_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gardener/controller-manager-library/pkg/fieldpath"
	"github.com/gardener/controller-manager-library/pkg/match"
)

var MatchFieldValue = match.MatchFieldValue
var MatchFieldPattern = match.MatchFieldPattern
var MatchNot = match.Not
var MatchOr = match.Or
var MatchAnd = match.And

type Nested struct {
	S2 string
	I2 int
}

type S struct {
	S1     string
	I1     int
	Nested Nested
}

var _ = Describe("Filter", func() {
	S1 := fieldpath.MustFieldPath(".S1")
	S2 := fieldpath.MustFieldPath(".Nested.S2")
	I1 := fieldpath.MustFieldPath(".I1")
	I2 := fieldpath.MustFieldPath(".Nested.I2")

	s := &S{
		S1: "string1",
		I1: 5,
		Nested: Nested{
			S2: "string2",
			I2: 15,
		},
	}
	Context("Matcher", func() {
		It("string field value", func() {
			Expect(MatchFieldValue(S1, "string1").Match(s)).To(BeTrue())
			Expect(MatchFieldValue(S1, "string2").Match(s)).To(BeFalse())
		})
		It("int field value", func() {
			Expect(MatchFieldValue(I1, 5).Match(s)).To(BeTrue())
			Expect(MatchFieldValue(I1, 15).Match(s)).To(BeFalse())
		})
		It("nested string field value", func() {
			Expect(MatchFieldValue(S2, "string2").Match(s)).To(BeTrue())
			Expect(MatchFieldValue(S2, "string1").Match(s)).To(BeFalse())
		})
		It("nested int field value", func() {
			Expect(MatchFieldValue(I2, 15).Match(s)).To(BeTrue())
			Expect(MatchFieldValue(I2, 5).Match(s)).To(BeFalse())
		})

		It("nested int field value (NOT)", func() {
			Expect(MatchNot(MatchFieldValue(I2, 15)).Match(s)).To(BeFalse())
			Expect(MatchNot(MatchFieldValue(I2, 5)).Match(s)).To(BeTrue())
		})

		It("int field value (AND)", func() {
			Expect(MatchAnd(MatchFieldValue(I2, 15), MatchFieldValue(I1, 5)).Match(s)).To(BeTrue())
			Expect(MatchAnd(MatchFieldValue(I2, 15), MatchFieldValue(I1, 15)).Match(s)).To(BeFalse())
			Expect(MatchAnd(MatchFieldValue(I2, 5), MatchFieldValue(I1, 5)).Match(s)).To(BeFalse())
		})

		It("int field value (OR)", func() {
			Expect(MatchOr(MatchFieldValue(I2, 15), MatchFieldValue(I1, 5)).Match(s)).To(BeTrue())
			Expect(MatchOr(MatchFieldValue(I2, 5), MatchFieldValue(I1, 5)).Match(s)).To(BeTrue())
			Expect(MatchOr(MatchFieldValue(I2, 15), MatchFieldValue(I1, 15)).Match(s)).To(BeTrue())
			Expect(MatchOr(MatchFieldValue(I2, 5), MatchFieldValue(I1, 55)).Match(s)).To(BeFalse())
		})

		Context("Filter", func() {
			list := []S{
				{
					S1: "alice",
					I1: 25,
				},
				{
					S1: "peter",
					I1: 26,
				},
				{
					S1: "bob",
					I1: 25,
				},
			}

			It("filters one", func() {
				Expect(match.FilterList(list, MatchFieldValue(S1, "peter"))).To(Equal(list[1:2]))
			})
			It("filters two", func() {
				Expect(match.FilterList(list, MatchFieldValue(I1, 25))).To(Equal([]S{list[0], list[2]}))
			})
			It("filters pattern", func() {
				Expect(match.FilterList(list, MatchFieldPattern(S1, ".*e.*"))).To(Equal([]S{list[0], list[1]}))
			})
		})
	})
})
