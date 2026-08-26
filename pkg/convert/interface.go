/*
 * SPDX-FileCopyrightText: Copyright Contributors to the Gardener project
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package convert

import (
	"github.com/gardener/controller-manager-library/pkg/utils"
)

// Interface maps nil struct pointer interfaces to nil.
// Interfaces not referring to an object will be converted to nil
func Interface(i interface{}) interface{} {
	if utils.IsNil(i) {
		return nil
	}
	return i
}
