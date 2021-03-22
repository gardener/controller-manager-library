/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package watches

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/extension"
)

////////////////////////////////////////////////////////////////////////////////
// Resource Flavors

// Conditional decribes a resource flavor checked if a dedicated contraint is met.
func Conditional(cond WatchConstraint, flavors ...ResourceFlavor) ResourceFlavor {
	return &conditional{
		cond:    cond,
		flavors: FlavoredResource(flavors),
	}
}

type conditional struct {
	cond    WatchConstraint
	flavors FlavoredResource
}

func (this *conditional) ResourceType(wctx WatchContext) ResourceKey {
	if this.cond == nil || this.cond.Check(wctx) {
		return this.flavors.ResourceKey(wctx)
	}
	return nil
}
func (this *conditional) HasResource(gk schema.GroupKind) bool {
	return this.flavors.HasResource(gk)
}
func (this *conditional) String() string {
	s := ""
	if this.cond != nil {
		s = fmt.Sprintf("if(%s)", this.cond)
	}
	return fmt.Sprintf("%s%s", s, this.flavors)
}

// ResourceFlavor is a single flavor valid for a dedicated constraint

func ResourceFlavorByGK(gk schema.GroupKind, constraints ...WatchConstraint) ResourceFlavor {
	return NewResourceFlavor(gk.Group, gk.Kind, constraints...)
}

func NewResourceFlavor(group, kind string, constraints ...WatchConstraint) ResourceFlavor {
	return &resourceFlavor{
		key:         extension.NewResourceKey(group, kind),
		constraints: constraints,
	}
}

type resourceFlavor struct {
	constraints []WatchConstraint
	key         ResourceKey
}

func (this *resourceFlavor) ResourceType(wctx WatchContext) ResourceKey {
	if len(this.constraints) == 0 {
		return this.key
	}
	for _, c := range this.constraints {
		if c.Check(wctx) {
			return this.key
		}
	}
	return nil
}
func (this *resourceFlavor) HasResource(gk schema.GroupKind) bool {
	return gk == this.key.GroupKind()
}

func (this *resourceFlavor) String() string {
	if len(this.constraints) == 0 {
		return this.key.String()
	}
	s := fmt.Sprintf("%s[", this.key)
	sep := ""
	for _, c := range this.constraints {
		s = fmt.Sprintf("%s%s%s", s, sep, c)
		sep = ", "
	}
	return s + "]"
}

////////////////////////////////////////////////////////////////////////////////
// utils

func SimpleResourceFlavors(group, kind string, constraints ...WatchConstraint) FlavoredResource {
	return FlavoredResource{NewResourceFlavor(group, kind, constraints...)}
}

func SimpleResourceFlavorsByKey(key ResourceKey, constraints ...WatchConstraint) FlavoredResource {
	return FlavoredResource{ResourceFlavorByGK(key.GroupKind(), constraints...)}
}

func SimpleResourceFlavorsByGK(gk schema.GroupKind, constraints ...WatchConstraint) FlavoredResource {
	return FlavoredResource{ResourceFlavorByGK(gk, constraints...)}
}
