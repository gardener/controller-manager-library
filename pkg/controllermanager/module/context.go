/*
 * SPDX-FileCopyrightText: Copyright Contributors to the Gardener project
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package module

import (
	"context"

	"github.com/gardener/controller-manager-library/pkg/ctxutil"
)

var ctx_module = ctxutil.NewValueKey(TYPE, (*module)(nil))

func GetModule(ctx context.Context) Interface {
	return ctx.Value(ctx_module).(Interface)
}
