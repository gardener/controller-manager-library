/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package bound

import (
	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/webhook"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/webhook/admission"
	handler "github.com/gardener/controller-manager-library/pkg/controllermanager/webhook/admission/bound"
	"github.com/gardener/controller-manager-library/pkg/logger"
)

func init() {
	webhook.Configure("test.gardener.cloud").
		Cluster(cluster.DEFAULT).
		Resource("core", "ResourceQuota").
		DefaultedStringOption("message", "yepp", "response message").
		Kind(admission.Validating(handler.Adapt(MyHandlerType))).
		MustRegister()
}

func MyHandlerType(webhook webhook.Interface) (handler.Interface, error) {
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

var _ handler.Interface = &MyHandler{}

func (this *MyHandler) Handle(logger logger.LogContext, req handler.Request) admission.Response {
	if req.Object != nil {
		logger.Infof("found bound object %s", req.Object.ObjectName())
	}
	return admission.Allowed(this.message)
	//return admission.Denied("aetsch")

}
