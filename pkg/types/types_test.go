/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 *
 */

package types_test

import (
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/gardener/controller-manager-library/pkg/types"
)

type xstring string
type xint int
type xfloat float32

type xlist []interface{}
type xxlist []xint
type xmap map[string]interface{}
type xxmap map[xstring]xint

var _ = Describe("Types test", func() {
	Context("Values", func() {
		It("normalizes", func() {

			m := xmap{
				"string": xstring("string"),
				"int":    xint(64),
				"float":  xfloat(3.25),
				"map": xxmap{
					"alice": 64,
				},
				"list":  xxlist{1, 2, 3},
				"list2": xlist{1},
			}

			r := map[string]interface{}{
				"string": "string",
				"int":    int64(64),
				"float":  float64(3.25),
				"map": map[string]interface{}{
					"alice": int64(64),
				},
				"list":  []interface{}{int64(1), int64(2), int64(3)},
				"list2": []interface{}{int64(1)},
			}

			n := types.CopyAndNormalize(m)
			Expect(n).NotTo(BeNil())

			reflect.DeepEqual(n, r)
			Expect(n).To(Equal(r))
		})
	})
})
