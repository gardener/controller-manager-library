/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package test

import (
	"fmt"
	"net/http"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/examples"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/server"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/server/handler"
)

func init() {
	server.Configure("test").AllowSecretMaintenance(true).
		RegisterHandler("demo", server.HandlerFunc("demo", http.HandlerFunc(demo))).
		RegisterHandler("message", Create).
		Register()
}

func demo(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("this is a test"))
}

type Handler struct {
	server server.Interface
	shared *examples.SharedInfo
}

func Create(server server.Interface) (handler.Interface, error) {
	return &Handler{
		server: server,
		shared: examples.GetShared(server.GetEnvironment()),
	}, nil
}

func (this *Handler) Setup() error {
	if this.shared != nil {
		this.server.Infof("found configured shared info")
		this.server.Register("message", this.handle)
	} else {
		this.server.Infof("no configured shared info found -> not serving message path")
	}
	return nil
}

func (this *Handler) handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/text")
	w.WriteHeader(http.StatusOK)
	m := this.shared.Values["message"]
	if m == nil {
		w.Write([]byte("no message configured\n"))
	}
	w.Write([]byte(fmt.Sprintf("%s\n", m)))
}
