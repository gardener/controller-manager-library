/*
 * SPDX-FileCopyrightText: Contributors to the Gardener project
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package server

import (
	"net/http"

	"github.com/gardener/controller-manager-library/pkg/logger"
)

var servMux = http.NewServeMux()

func Register(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	logger.InitInfof("adding %s endpoint", pattern)
	servMux.HandleFunc(pattern, handler)
}

func RegisterHandler(pattern string, handler http.Handler) {
	logger.InitInfof("adding %s endpoint", pattern)
	servMux.Handle(pattern, handler)
}
