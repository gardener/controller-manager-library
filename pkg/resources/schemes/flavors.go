/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package schemes

import (
	"fmt"
	"reflect"
	runtime2 "runtime"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/gardener/controller-manager-library/pkg/resources"
)

////////////////////////////////////////////////////////////////////////////////
// Scheme Flavors

////////////////////////////////////////////////////////////////////////////////

func ExplicitSchemeSource(scheme *runtime.Scheme, desc ...string) SchemeSource {
	if len(desc) == 0 {
		desc = []string{"explicit"}
	}
	return &schemeFunction{
		function: SchemeFunction(func() *runtime.Scheme { return scheme }),
		desc:     strings.Join(desc, " "),
	}
}

func SchemeFunctionSource(scheme func() *runtime.Scheme) SchemeSource {
	val := reflect.ValueOf(scheme)
	return &schemeFunction{
		function: SchemeFunction(scheme),
		desc:     runtime2.FuncForPC(val.Pointer()).Name(),
	}
}

// Conditional decribes a resource flavor checked if a dedicated contraint is met.
func Conditional(cond SchemeConstraint, flavors ...SchemeSource) SchemeSource {
	return &conditional{
		cond:    cond,
		flavors: FlavoredSchemeSource(flavors),
	}
}

type conditional struct {
	cond    SchemeConstraint
	flavors FlavoredSchemeSource
}

func (this *conditional) Equivalent(o SchemeSource) bool {
	return this == o
}

func (this *conditional) Scheme(infos *resources.ResourceInfos) *runtime.Scheme {
	if this.cond == nil || this.cond.Check(infos) {
		return this.flavors.Scheme(infos)
	}
	return nil
}

func (this *conditional) String() string {
	s := ""
	if this.cond != nil {
		s = fmt.Sprintf("IF(%s)", this.cond)
	}
	return fmt.Sprintf("%s%s", s, this.flavors)
}
