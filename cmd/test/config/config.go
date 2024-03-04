/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"fmt"

	"github.com/gardener/controller-manager-library/pkg/config"
	"github.com/spf13/pflag"
)

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

/*
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
*/
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

	fmt.Print(flags.FlagUsages())
	fmt.Printf("setting args\n")
	_ = flags.Set("a", "test")

	_ = main.Evaluate()
	fmt.Printf("main.a   = %v\n", main.GetOption("a").Value())
	fmt.Printf("bob.a    = %v\n", bob.GetOption("a").Value())
	fmt.Printf("alice.a  = %v\n", alice.GetOption("a").Value())
}

func ConfigMain() {
	doB()
}
