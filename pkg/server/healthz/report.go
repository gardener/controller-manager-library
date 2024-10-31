/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package healthz

import (
	"fmt"
	"sync"
	"time"

	"github.com/gardener/controller-manager-library/pkg/logger"
)

func Tick(key string) {
	lock.Lock()
	defer lock.Unlock()

	if c := checks[key]; c != nil {
		c.last = time.Now()
	} else {
		logger.Warnf("check with key %q not configured", key)
	}
}

func Start(key string, period time.Duration) {
	lock.Lock()
	defer lock.Unlock()

	checks[key] = &check{time.Now(), 3 * period}
}

func End(key string) {
	lock.Lock()
	defer lock.Unlock()

	removeCheck(key)
}

type check struct {
	last    time.Time
	timeout time.Duration
}

var (
	checks = map[string]*check{}
	lock   sync.Mutex
)

func removeCheck(key string) {
	delete(checks, key)
}

func IsHealthy() bool {
	lock.Lock()
	defer lock.Unlock()

	now := time.Now()

	for key, c := range checks {
		limit := now.Add(-c.timeout)
		if c.last.Before(limit) {
			logger.Warnf("outdated health check '%s': %s", key, limit.Sub(c.last))
			return false
		}
		logger.Debugf("%s: %s", key, c.last)
	}
	return true
}

func HealthInfo() (bool, string) {
	lock.Lock()
	defer lock.Unlock()

	info := ""
	now := time.Now()
	for key, c := range checks {
		limit := now.Add(-c.timeout)
		info = fmt.Sprintf("%s%s: %s\n", info, key, c.last)
		if c.last.Before(limit) {
			logger.Warnf("outdated health check '%s': %s", key, limit.Sub(c.last))
			return false, info
		}
		logger.Debugf("%s: %s", key, c.last)
	}
	return true, info
}
