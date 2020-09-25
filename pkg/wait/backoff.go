/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package wait

import (
	"context"
	"errors"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

type ConditionFunc = wait.ConditionFunc
type Backoff = wait.Backoff

var ErrWaitTimeout = wait.ErrWaitTimeout
var ErrWaitCanceled = errors.New("backoff cancelled")

func ExponentialBackoff(ctx context.Context, backoff Backoff, condition ConditionFunc) error {
	var timer *time.Timer
	var endless = backoff.Steps < 0
	for backoff.Steps > 0 || endless {
		if endless {
			backoff.Steps = 2
		}
		if ok, err := condition(); err != nil || ok {
			return err
		}
		if backoff.Steps == 1 {
			break
		}

		if timer == nil {
			timer = time.NewTimer(backoff.Step())
		} else {
			timer.Reset(backoff.Step())
		}
		select {
		case <-timer.C:
		case <-ctx.Done():
			return ErrWaitCanceled

		}
	}
	return ErrWaitTimeout
}
