/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
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
	Start()
}

func Start(ctx context.Context, startInterface StartInterface, synched ...cache.InformerSynced) error {
	startInterface.Start()
	if ok := cache.WaitForCacheSync(ctx.Done(), synched...); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}
	return nil
}
