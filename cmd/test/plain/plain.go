/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
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
