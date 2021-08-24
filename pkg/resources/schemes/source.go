/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package schemes

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/gardener/controller-manager-library/pkg/resources"
)

type SchemeSource = resources.SchemeSource

// SchemeFunction is a function providing a Scheme
type SchemeFunction func() *runtime.Scheme

type schemeFunction struct {
	function SchemeFunction
	desc     string
}

func (this *schemeFunction) Equivalent(o SchemeSource) bool {
	if this == o {
		return true
	}
	d, ok := o.(*schemeFunction)
	if !ok {
		return false
	}
	return d.function() == this.function()
}

func (this *schemeFunction) Scheme(infos *resources.ResourceInfos) *runtime.Scheme {
	return this.function()
}

func (this *schemeFunction) String() string {
	return this.desc
}

////////////////////////////////////////////////////////////////////////////////

type FlavoredSchemeSource []SchemeSource

func NewFlavoredSchemeSource(flavors ...SchemeSource) FlavoredSchemeSource {
	return flavors
}

func (this FlavoredSchemeSource) Scheme(infos *resources.ResourceInfos) *runtime.Scheme {
	var s *runtime.Scheme
	for _, f := range this {
		s = f.Scheme(infos)
		if s != nil {
			break
		}
	}
	return s
}

func (this FlavoredSchemeSource) Equivalent(o SchemeSource) bool {
	s, ok := o.(FlavoredSchemeSource)
	if !ok {
		return false
	}
	if len(this) != len(s) {
		return false
	}
	for i, f := range this {
		if !f.Equivalent(s[i]) {
			return false
		}
	}
	return true
}

func (this FlavoredSchemeSource) String() string {
	s := "["
	sep := ""
	for _, f := range this {
		s = fmt.Sprintf("%s%s%s", s, sep, f)
		sep = ", "
	}
	return s + "]"
}
