/*
 * Copyright 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
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
