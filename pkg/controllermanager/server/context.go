/*
 * SPDX-FileCopyrightText: Contributors to the Gardener project
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package server

import (
	"context"

	"github.com/gardener/controller-manager-library/pkg/ctxutil"
)

var ctx_server = ctxutil.NewValueKey(TYPE, (*httpserver)(nil))

func GetServer(ctx context.Context) Interface {
	return ctx.Value(ctx_server).(Interface)
}
