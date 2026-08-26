/*
 * SPDX-FileCopyrightText: Copyright Contributors to the Gardener project
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package webhook

import (
	"context"

	"github.com/gardener/controller-manager-library/pkg/ctxutil"
)

var ctx_webhook = ctxutil.NewValueKey(TYPE, (*webhook)(nil))

func GetWebhook(ctx context.Context) Interface {
	return ctx_webhook.Get(ctx).(Interface)
}
