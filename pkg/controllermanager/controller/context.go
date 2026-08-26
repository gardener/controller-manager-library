/*
 * SPDX-FileCopyrightText: Copyright Contributors to the Gardener project
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"

	"github.com/gardener/controller-manager-library/pkg/ctxutil"
)

var ctx_controller = ctxutil.NewValueKey(TYPE, (*controller)(nil))

func GetController(ctx context.Context) Interface {
	return ctx.Value(ctx_controller).(Interface)
}
