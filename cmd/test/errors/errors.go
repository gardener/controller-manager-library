/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package errors

import (
	"fmt"
	"github.com/gardener/controller-manager-library/pkg/errors"
)

const GRP = "Group1"

var (
	Err1 = errors.DeclareFormalType(GRP, "Err1", "%s: %s")
)

func Show(msg string, err error, print bool) {
	fmt.Printf("%s: %s\n", msg, err)
	if c, ok := err.(errors.Categorized); ok {
		fmt.Printf("  group: %s\n", c.Group())
		fmt.Printf("  kind : %s\n", c.Kind())
	}
	if c, ok := err.(errors.Formal); ok {
		for i := 0; i < c.Length(); i++ {
			fmt.Printf("  arg %d: %v\n", i, c.Arg(i))
		}
	}
	if c, ok := err.(errors.StackTracer); ok {
		fmt.Printf("StackTrace:")
		fmt.Printf("%+v\n", c.StackTrace())
	}
	if c, ok := err.(errors.Categorized); ok {
		if c.Cause() != nil {
			Show("Nested", c.Cause(), false)
		}
	}
	if print {
		fmt.Printf(" s: %s\n", err)
		fmt.Printf(" q: %q\n", err)
		fmt.Printf(" v: %v\n", err)
		fmt.Printf("-v: %-v\n", err)
		fmt.Printf("+v: %+v\n", err)
	}
}

func ErrorsMain() {

	err := Err1.New("object1", "invalid")

	Show("Err1", err, true)

	err2 := Err1.Wrap(err, "object2", "pending")
	fmt.Printf("*********************************************\n")
	Show("Err2", err2, true)
}
