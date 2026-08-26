/*
 * SPDX-FileCopyrightText: Contributors to the Gardener project
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package ctxutil

import (
	"context"
)

var cancelkey = ""

func CancelContext(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	return context.WithValue(ctx, &cancelkey, cancel)
}

func Cancel(ctx context.Context) {
	ctx.Value(&cancelkey).(context.CancelFunc)()
}
