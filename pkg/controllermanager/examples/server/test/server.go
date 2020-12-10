/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package test

import (
	"net/http"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/server"
)

func init() {
	server.Configure("test").
		RegisterHandler("demo", server.HandlerFunc("demo", http.HandlerFunc(demo))).
		Register()
}

func demo(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("this is a test"))
}
