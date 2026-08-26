/*
 * SPDX-FileCopyrightText: Contributors to the Gardener project
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package controllermanager

import (
	"fmt"
)

type Mapping interface {
	Map(name string) string
}

type DefaultMapping map[string]string

func (this DefaultMapping) Map(n string) string {
	return this[n]
}

func (this DefaultMapping) String() string {
	r := "{"
	sep := ""
	for s, d := range this {
		r = fmt.Sprintf("%s%s%s=>%s", r, sep, s, d)
		sep = ", "
	}
	return r + "}"
}
