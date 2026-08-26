/*
 * SPDX-FileCopyrightText: Contributors to the Gardener project
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package reconcile

import (
	"time"

	"github.com/gardener/controller-manager-library/pkg/logger"
)

func AbortAndDelayOnError(logger logger.LogContext, err error) {
	if err != nil {
		panic(Delay(logger, err))
	}
}

func AbortAndRepeatOnError(logger logger.LogContext, err error) {
	if err != nil {
		panic(Repeat(logger, err))
	}
}

func AbortAndRecheckOnError(logger logger.LogContext, err error, interval ...time.Duration) {
	if err != nil {
		panic(Recheck(logger, err, interval...))
	}
}

func AbortWith(status Status) {
	panic(status)
}
