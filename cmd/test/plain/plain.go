/*
 * Copyright 2019 SAP SE or an SAP affiliate company. All rights reserved.
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

package plain

import (
	"context"
	"fmt"
	"github.com/gardener/controller-manager-library/pkg/resources/plain"
	"k8s.io/api/core/v1"
)

var serviceAccountYAML = `
  apiVersion: v1
  kind: ServiceAccount
  metadata:
    labels:
      app: dns
    name: dns
    namespace: kube-system
`

func PlainMain() {
	ctx := plain.NewResourceContext(context.TODO(), plain.DefaultScheme())

	resources := ctx.Resources()
	obj, err := resources.Decode([]byte(serviceAccountYAML))
	if err != nil {
		fmt.Printf("decode failed: %s\n", err)
		return
	}

	resources.Resources()
	resources.ResourceContext()
	fmt.Printf("type: %s, sa: %t, gvk: %s\n", obj.GetResource().ObjectType(), obj.IsA(v1.ServiceAccount{}), obj.GroupVersionKind())
}
