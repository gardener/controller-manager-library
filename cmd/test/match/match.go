// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package match

import (
	"fmt"
	"github.com/gardener/controller-manager-library/pkg/utils"
)

func MatchMain() {
	fmt.Printf("pattern matching\n")
	g := utils.NewStringGlobMatcher("abc")
	test(g, "abc", true)
	test(g, "abd", false)
	test(g, "bbc", false)

	g = utils.NewStringGlobMatcher("a*b?c")

	test(g, "abbc", true)
	test(g, "abc", false)

	g = utils.NewStringGlobMatcher("a*")
	test(g, "abc", true)
	test(g, "bbc", false)
	test(g, "b", false)
	test(g, "", false)

	p := utils.NewPathGlobMatcher("alice/**/bob/*/tom")
	testP(p, "alice/bob/bob/tom", true)
	testP(p, "alive/bob/tom", false)

	p = utils.NewPathGlobMatcher("alice/**/b*b/*/tom")
	testP(p, "alice/bob/bob/tom", true)
	testP(p, "alice/bla/blub/bob/any/tom", true)
	testP(p, "alice/bla/blub/any/tom", true)
	testP(p, "alice/bla/any/tom", false)
	testP(p, "alice/bob/tom", false)

}

func test(g utils.Matcher, cand string, expected bool) {
	r := g.Match(cand)
	if r != expected {
		fmt.Printf("*** %s:  %s expected %t\n", cand, g, expected)
	}
}
func testP(g utils.Matcher, cand string, expected bool) {
	r := g.Match(cand)
	if r != expected {
		fmt.Printf("*** %s:  %s expected %t\n", cand, g, expected)
	}
}
