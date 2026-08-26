/*
 * SPDX-FileCopyrightText: Copyright Contributors to the Gardener project
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package informerfactories

import (
	"context"
	"fmt"

	"k8s.io/client-go/tools/cache"
)

type StartInterface interface {
	Start(<-chan struct{})
}

func Start(ctx context.Context, startInterface StartInterface, synched ...cache.InformerSynced) error {
	startInterface.Start(ctx.Done())
	if ok := cache.WaitForCacheSync(ctx.Done(), synched...); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}
	return nil
}
