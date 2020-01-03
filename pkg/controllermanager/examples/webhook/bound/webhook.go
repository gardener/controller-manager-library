/*
 * Copyright 2020 SAP SE or an SAP affiliate company. All rights reserved.
 * This file is licensed under the Apache Software License, v. 2 except as noted
 * otherwise in the LICENSE file
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 *
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
		Handler(handler.Adapt(MyHandlerType)).
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
	return admission.Denied("aetsch")

}
