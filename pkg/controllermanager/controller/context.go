/*
 * Copyright 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package controller

import (
	"context"
	"github.com/gardener/controller-manager-library/pkg/utils"
	"reflect"
)

var controller_key reflect.Type
var extension_key reflect.Type

func init() {
	controller_key, _ = utils.TypeKey((*controller)(nil))
}

func setController(ctx context.Context, c *controller) context.Context {
	return context.WithValue(ctx, controller_key, c)
}

func GetController(ctx context.Context) Interface {
	return ctx.Value(controller_key).(Interface)
}

func setExtension(ctx context.Context, e *Extension) context.Context {
	return context.WithValue(ctx, extension_key, e)
}

func GetExtension(ctx context.Context) *Extension {
	return ctx.Value(extension_key).(*Extension)
}
