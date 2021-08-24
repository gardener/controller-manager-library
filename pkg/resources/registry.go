/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/gardener/controller-manager-library/pkg/resources/abstract"
)

func DeclareDefaultVersion(gv schema.GroupVersion) {
	abstract.DeclareDefaultVersion(gv)
}

func DefaultVersion(g string) string {
	return abstract.DefaultVersion(g)
}

func Register(builders ...runtime.SchemeBuilder) {
	abstract.Register(builders...)
}

func DefaultScheme() *runtime.Scheme {
	return abstract.DefaultScheme()
}

var defaultSchemeSource SchemeSource

func SetDefaultSchemeSource(src SchemeSource) {
	defaultSchemeSource = src
}

func DefaultSchemeSource() SchemeSource {
	if defaultSchemeSource != nil {
		return defaultSchemeSource
	}
	return defaultScheme
}

var defaultScheme = &_defaultScheme{DefaultScheme()}

type _defaultScheme struct {
	scheme *runtime.Scheme
}

func (this *_defaultScheme) Equivalent(o SchemeSource) bool {
	if this == o {
		return true
	}
	d, ok := o.(*_defaultScheme)
	if !ok {
		return false
	}
	return d.scheme == this.scheme
}

func (this *_defaultScheme) Scheme(infos *ResourceInfos) *runtime.Scheme {
	return this.scheme
}

func (this *_defaultScheme) String() string {
	return "default"
}
