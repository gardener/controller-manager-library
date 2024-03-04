/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"github.com/gardener/controller-manager-library/pkg/logger"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/webhook"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/webhook/admission"
)

func init() {
	webhook.Configure("test.gardener.cloud").
		Kind(admission.Validating(MyHandlerType)).
		Cluster(webhook.CLUSTER_MAIN).
		Resource("core", "ResourceQuota").
		DefaultedStringOption("message", "yepp", "response message").
		MustRegister()
}

func MyHandlerType(webhook webhook.Interface) (admission.Interface, error) {
	msg, err := webhook.GetStringOption("message")
	if err == nil {
		webhook.Infof("found option message: %s", msg)
	}
	return &MyHandler{message: msg, hook: webhook}, nil
}

type MyHandler struct {
	message string
	admission.DefaultHandler
	hook webhook.Interface
}

var _ admission.Interface = &MyHandler{}

func (this *MyHandler) Handle(logger.LogContext, admission.Request) admission.Response {
	return admission.Allowed(this.message)
	//return admission.Denied("aetsch")
}
