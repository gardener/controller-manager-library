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

package config_test

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"

	"github.com/gardener/controller-manager-library/pkg/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	var (
		main  *config.DefaultOptionSet
		flags *pflag.FlagSet
	)

	BeforeEach(func() {
		main = config.NewDefaultOptionSet("main", "")
		flags = pflag.NewFlagSet("test", pflag.ExitOnError)
	})

	It("should handle StringOption", func() {
		var a, b, c, d string
		main.AddStringOption(&a, "maina", "a", "A", "maina name")
		main.AddStringOption(&b, "mainb", "b", "B", "mainb name")
		main.AddStringOption(&c, "mainc", "c", "C", "mainc name")
		pd := main.AddStringOption(&d, "maind", "d", "D", "maind name")
		Expect(pd).To(Equal(&d))
		pe := main.AddStringOption(nil, "maine", "e", "E", "maine name")
		Expect(pe).NotTo(BeNil())
		main.AddStringOption(nil, "mainf", "", "", "maind name")
		main.AddToFlags(flags)

		err := flags.Parse([]string{"--maina", "va", "-b", "vb", "--maind=vd", "-e=ve"})
		err = main.Evaluate()
		Expect(err).NotTo(HaveOccurred())

		Expect(a).To(Equal("va"))
		Expect(main.GetOption("maina").StringValue()).To(Equal("va"))
		Expect(main.GetOption("maina").Description).To(Equal("maina name"))
		Expect(main.GetOption("maina").Changed()).To(Equal(true))

		Expect(b).To(Equal("vb"))
		Expect(main.GetOption("mainb").StringValue()).To(Equal("vb"))

		Expect(c).To(Equal("C"))
		Expect(main.GetOption("mainc").StringValue()).To(Equal("C"))
		Expect(main.GetOption("mainc").Changed()).To(Equal(false))

		Expect(d).To(Equal("vd"))
		Expect(main.GetOption("maind").StringValue()).To(Equal("vd"))

		Expect(*pe).To(Equal("ve"))
		Expect(main.GetOption("maine").StringValue()).To(Equal("ve"))

		Expect(main.GetOption("mainf").StringValue()).To(Equal(""))
	})

	It("should handle IntOption", func() {
		main.AddIntOption(nil, "maina", "a", 1, "maina name")
		pb := main.AddIntOption(nil, "mainb", "b", 1, "mainb name")
		main.AddToFlags(flags)

		err := flags.Parse([]string{"--maina", "-123"})
		err = main.Evaluate()
		Expect(err).NotTo(HaveOccurred())

		Expect(main.GetOption("maina").IntValue()).To(Equal(-123))
		Expect(main.GetOption("mainb").IntValue()).To(Equal(1))
		Expect(*pb).To(Equal(1))
	})

	It("should handle UintOption", func() {
		main.AddUintOption(nil, "maina", "a", 1, "maina name")
		pb := main.AddUintOption(nil, "mainb", "b", 1, "mainb name")
		main.AddToFlags(flags)

		err := flags.Parse([]string{"-a", "123"})
		err = main.Evaluate()
		Expect(err).NotTo(HaveOccurred())

		Expect(main.GetOption("maina").StringValue()).To(Equal(""))
		Expect(main.GetOption("maina").IntValue()).To(Equal(0))
		Expect(main.GetOption("maina").UintValue()).To(Equal(uint(123)))
		Expect(main.GetOption("mainb").UintValue()).To(Equal(uint(1)))
		Expect(*pb).To(Equal(uint(1)))
	})

	It("should handle BoolOption", func() {
		main.AddBoolOption(nil, "maina", "a", false, "maina name")
		main.AddBoolOption(nil, "mainb", "b", true, "mainb name")
		pc := main.AddBoolOption(nil, "mainc", "c", false, "mainc name")
		main.AddBoolOption(nil, "maind", "d", false, "maind name")
		main.AddToFlags(flags)

		err := flags.Parse([]string{"-ac=true"})
		err = main.Evaluate()
		Expect(err).NotTo(HaveOccurred())

		Expect(main.GetOption("maina").BoolValue()).To(Equal(true))
		Expect(main.GetOption("mainb").BoolValue()).To(Equal(true))
		Expect(main.GetOption("mainc").BoolValue()).To(Equal(true))
		Expect(main.GetOption("maind").BoolValue()).To(Equal(false))
		Expect(*pc).To(Equal(true))
	})

	It("should handle DurationOption", func() {
		main.AddDurationOption(nil, "maina", "a", 1*time.Second, "maina name")
		pb := main.AddDurationOption(nil, "mainb", "b", 1*time.Second, "mainb name")
		main.AddToFlags(flags)

		err := flags.Parse([]string{"-a", "3h"})
		err = main.Evaluate()
		Expect(err).NotTo(HaveOccurred())

		Expect(main.GetOption("maina").DurationValue()).To(Equal(3 * time.Hour))
		Expect(main.GetOption("mainb").DurationValue()).To(Equal(1 * time.Second))
		Expect(*pb).To(Equal(1 * time.Second))
	})

	It("should handle StringArrayOption", func() {
		main.AddStringArrayOption(nil, "maina", "a", []string{"q", "w"}, "maina name")
		pb := main.AddStringArrayOption(nil, "mainb", "b", []string{"q", "w"}, "mainb name")
		main.AddStringArrayOption(nil, "mainc", "c", []string{"q", "w"}, "mainc name")
		main.AddStringArrayOption(nil, "maind", "d", []string{}, "maind name")
		main.AddToFlags(flags)

		err := flags.Parse([]string{"-a", "xx", "-a", "yy", "-b=eee,fff"})
		err = main.Evaluate()
		Expect(err).NotTo(HaveOccurred())

		Expect(main.GetOption("maina").StringArray()).To(Equal([]string{"xx", "yy"}))
		Expect(main.GetOption("maina").IsArray()).To(Equal(true))
		Expect(main.GetOption("maina").IsSlice()).To(Equal(false))

		Expect(main.GetOption("mainb").StringArray()).To(Equal([]string{"eee,fff"}))
		Expect(main.GetOption("mainb").IsArray()).To(Equal(true))
		Expect(main.GetOption("mainb").IsSlice()).To(Equal(false))
		Expect(main.GetOption("mainb").Changed()).To(Equal(true))
		Expect(*pb).To(Equal([]string{"eee,fff"}))

		Expect(main.GetOption("mainc").StringArray()).To(Equal([]string{"q", "w"}))
		Expect(main.GetOption("mainc").Changed()).To(Equal(false))
		Expect(main.GetOption("maind").StringArray()).To(Equal([]string{}))
	})

	It("should handle shared and generic options", func() {
		main.AddStringOption(nil, "main", "m", "def-main", "main name")

		shared := config.NewSharedOptionSet("", "controller.test", nil)
		shared.AddIntOption(nil, "size", "s", 3, "pool size")
		shared.AddIntOption(nil, "cnt", "c", 1, "worker count")
		shared.AddIntOption(nil, "bar", "b", 55, "bar name")
		main.AddSource(shared.Name(), shared)

		shared2 := config.NewSharedOptionSet("", "controller.foo", nil)
		shared2.AddIntOption(nil, "size", "s", 3, "pool size")
		shared2.AddIntOption(nil, "cnt", "c", 1, "worker count")
		shared2.Unshare("bar")
		main.AddSource(shared2.Name(), shared2)

		generic := config.NewGenericOptionSource("generic", "generic", func(s string) string { return s + " for generic" })
		generic.AddStringOption(config.Flat, nil, "flat", "", "def-flat", "flat option")
		generic.AddStringOption(config.Prefixed, nil, "prefixed", "", "def-prefixed", "prefixed option")
		generic.AddStringOption(config.PrefixedShared, nil, "prefixshared", "", "def-prefixshared", "prefix shared option")
		generic.AddStringOption(config.Shared, nil, "shared", "", "def-shared", "shared option")
		main.AddSource(generic.Name(), generic)

		prefixed := config.NewDefaultOptionSet("pool", "pool")
		generic.PrefixedShared().AddSource("pool", prefixed)
		prefixed.AddIntOption(nil, "size", "", 1, "pool size")

		main.AddToFlags(flags)

		err := flags.Parse([]string{"--main", "vm", "-s", "4", "--controller.test.size=5", "--prefixshared=a", "--generic.prefixshared=b"})
		err = main.Evaluate()
		Expect(err).NotTo(HaveOccurred())

		var actual []string
		main.VisitOptions(func(o *config.ArbitraryOption) bool {
			actual = append(actual, fmt.Sprintf("%s: %t: %v (%s)", o.Name, o.Changed(), o.Value(), o.Description))
			return true
		})

		Expect(actual).To(
			ConsistOf(
				"bar: false: 0 (bar name)",
				"cnt: false: 0 (worker count)",
				"controller.foo.cnt: false: 1 (worker count)",
				"controller.foo.size: false: 4 (pool size)",
				"controller.test.bar: false: 55 (bar name)",
				"controller.test.cnt: false: 1 (worker count)",
				"controller.test.size: true: 5 (pool size)",
				"flat: false: def-flat (flat option)",
				"generic.pool.size: false: 1 (pool size for generic)",
				"generic.prefixed: false: def-prefixed (prefixed option)",
				"generic.prefixshared: true: b (prefix shared option for generic)",
				"main: true: vm (main name)",
				"pool.size: false: 0 (pool size)",
				"prefixshared: true: a (prefix shared option)",
				"shared: false:  (shared option)",
				"size: true: 4 (pool size)",
			),
		)
	})
})
