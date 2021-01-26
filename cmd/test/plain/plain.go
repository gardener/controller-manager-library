/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package plain

import (
	"context"
	"fmt"

	"k8s.io/api/core/v1"

	"github.com/gardener/controller-manager-library/pkg/resources/plain"
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

type X struct {
	A string
}

type Y X

func (y *Y) Do() {}

func PlainMain() {
	x := &X{
		"test",
	}

	y := (*Y)(x)
	fmt.Printf("%t\n", (*Y)(x) == y)
	y.Do()

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
