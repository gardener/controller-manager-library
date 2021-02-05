/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"fmt"
	"os"
	"reflect"

	"github.com/ghodss/yaml"
	"github.com/spf13/pflag"

	"github.com/gardener/controller-manager-library/pkg/config"
)

const configData = `
  size: 5
  bool: true
  main: 
    - not changed
    - changed
  controller:
    test:
      cnt: 4
`

type Mine struct {
	Option string `configmain:"option,'dies ist ein test'"`
}

type Targets struct {
	data string
	test string
}

var _ config.OptionSource = &Targets{}

func (t *Targets) AddOptionsToSet(set config.OptionSet) {
	set.AddStringOption(&t.data, "data", "d", "none", "test data")
	set.AddStringOption(&t.test, "test", "", "none", "test name")
}

func doA() {
	var data map[string]interface{}
	err := yaml.Unmarshal([]byte(configData), &data)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	args, err := config.MapToArguments("", nil, data)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	flags := pflag.NewFlagSet("map", pflag.ExitOnError)
	flags.Bool("bool", false, "count")
	flags.Int("controller.test.cnt", 1, "count")
	flags.Int("size", 1, "count")
	flags.String("main", "", "count")
	flags.Set("controller.test.cnt", "20")
	config.MergeFlags(flags, args, true)
	flags.VisitAll(func(flag *pflag.Flag) {
		fmt.Printf("eff %s: %v\n", flag.Name, flag.Value)
	})

	fmt.Printf("%d args\n", len(args))
	for i, a := range args {
		fmt.Printf("%d: %s\n", i, a)
	}
	main := config.NewDefaultOptionSet("configmain", "")
	main.AddStringOption(nil, "main", "m", "main", "main name")

	shared := config.NewSharedOptionSet("", "controller.test", nil)
	shared.AddIntOption(nil, "size", "s", 3, "pool size")
	shared.AddIntOption(nil, "cnt", "c", 1, "worker count")
	main.AddSource(shared.Name(), shared)

	targets := &Targets{}
	main.AddSource("targets", targets)

	generic := config.NewGenericOptionSource("generic", "generic", func(s string) string { return s + " for generic" })
	main.AddSource("generic", generic)

	generic.AddStringOption(config.Flat, nil, "generic", "", "yes", "generic option")
	generic.AddStringOption(config.Prefixed, nil, "prefixed", "", "yes", "prefixed")
	generic.AddStringOption(config.PrefixedShared, nil, "shared", "", "shared", "shared name")
	generic.AddStringOption(config.Shared, nil, "main", "m", "main", "main name")

	prefixed := config.NewDefaultOptionSet("pool", "pool")
	generic.PrefixedShared().AddSource("pool", prefixed)
	prefixed.AddUintOption(nil, "size", "", 1, "pool size")

	fmt.Printf("adding args to command line\n")

	flags = pflag.NewFlagSet("test", pflag.ExitOnError)
	main.AddToFlags(flags)

	fmt.Printf("setting args\n")
	flags.Set("main", "changed")
	flags.Set("size", "5")
	flags.Set("controller.test.cnt", "4")
	flags.Set("test", "4")
	flags.Set("shared", "globallychanged")

	fmt.Printf("evaluate args\n")
	main.Evaluate()
	fmt.Printf("print args\n")
	config.Print(config.PrintfWriter, "", main)
	fmt.Printf("targets: %#v\n", targets)
}

func evaluate(o interface{}) {
	t := reflect.TypeOf(o)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		fmt.Printf("No struct\n")
		return
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		f.Tag.Lookup("configmain")
		fmt.Printf("%s: %s:  %s\n", f.Name, f.Type, f.Tag.Get("configmain"))
	}
}

func doB() {
	alice := config.NewSharedOptionSet("alice", "")
	bob := config.NewSharedOptionSet("bob", "")

	var s1 string

	alice.AddStringOption(&s1, "a", "", "", "option alice a")
	bob.AddStringOption(&s1, "a", "", "", "option bob a")
	alice.AddSource("s-bob", bob)

	main := config.NewDefaultOptionSet("alice", "")
	main.AddSource("s-alice", alice)

	flags := pflag.NewFlagSet("test", pflag.ExitOnError)
	main.AddToFlags(flags)

	fmt.Printf(flags.FlagUsages())
	fmt.Printf("setting args\n")
	flags.Set("a", "test")

	main.Evaluate()
	fmt.Printf("main.a   = %v\n", main.GetOption("a").Value())
	fmt.Printf("bob.a    = %v\n", bob.GetOption("a").Value())
	fmt.Printf("alice.a  = %v\n", alice.GetOption("a").Value())
}

func ConfigMain() {
	doB()
}
