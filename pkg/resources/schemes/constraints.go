/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package schemes

import (
	"fmt"

	"github.com/Masterminds/semver/v3"

	"github.com/gardener/controller-manager-library/pkg/resources"
)

// Constraints for selecting schemes for a dedicated cluster
// A constraint is always based on a
type SchemeConstraint interface {
	Check(infos *resources.ResourceInfos) bool
}

// ConstraintFunction maps a check function to a constraint interface
type ConstraintFunction func(infos *resources.ResourceInfos) bool

func (this ConstraintFunction) Check(infos *resources.ResourceInfos) bool {
	return this(infos)
}

type constraintFunction struct {
	ConstraintFunction
	desc string
}

func (this *constraintFunction) String() string {
	return this.desc
}

// Not negates a constraint
func Not(c SchemeConstraint) SchemeConstraint {
	return &constraintFunction{
		ConstraintFunction: func(infos *resources.ResourceInfos) bool {
			return !c.Check(infos)
		},
		desc: fmt.Sprintf("NOT(%s)", c),
	}
}

// And checks multiple constraints to be true
func And(c ...SchemeConstraint) SchemeConstraint {
	desc := "AND("
	sep := ""
	for _, e := range c {
		desc = fmt.Sprintf("%s%s%s", desc, sep, e)
		sep = ", "
	}
	return &constraintFunction{
		ConstraintFunction: func(infos *resources.ResourceInfos) bool {
			for _, e := range c {
				if !e.Check(infos) {
					return false
				}
			}
			return true
		},
		desc: desc + ")",
	}
}

// Or checks multiple constraints to be not false
func Or(c ...SchemeConstraint) SchemeConstraint {
	desc := "OR("
	sep := ""
	for _, e := range c {
		desc = fmt.Sprintf("%s%s%s", desc, sep, e)
		sep = ", "
	}
	return &constraintFunction{
		ConstraintFunction: func(infos *resources.ResourceInfos) bool {
			for _, e := range c {
				if e.Check(infos) {
					return true
				}
			}
			return false
		},
		desc: desc + ")",
	}
}

func APIServerVersion(constraint string) SchemeConstraint {
	c, err := semver.NewConstraint(constraint)
	if err != nil {
		panic(err)
	}
	return &constraintFunction{
		ConstraintFunction: func(infos *resources.ResourceInfos) bool {
			return infos != nil && c.Check(infos.GetServerVersion())
		},
		desc: fmt.Sprintf("(server version %s)", constraint),
	}
}
