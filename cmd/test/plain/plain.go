/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package plain

import (
	"context"
	"fmt"
	"sync"

	"k8s.io/api/core/v1"

	"github.com/gardener/controller-manager-library/pkg/goutils"
	"github.com/gardener/controller-manager-library/pkg/resources/plain"
	"github.com/gardener/controller-manager-library/pkg/utils"
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

///////////////////////
type X struct {
	A string
}

type Y X

func (y *Y) Do() {}

func doX() {
	x := &X{
		"test",
	}

	y := (*Y)(x)
	fmt.Printf("%t\n", (*Y)(x) == y)
	y.Do()
}

////////////////////////

type A interface {
	Do()
}

type AS struct {
	DoFunc func()
}

func (a AS) Do() {
	if a.DoFunc != nil {
		a.DoFunc()
	}
}

func doA() {
	m := map[A]struct{}{}
	as := AS{
		nil,
	}

	var a A = as
	var b A = a
	var c A = &as

	if utils.IsComparable(a) {
		m[a] = struct{}{}
		fmt.Printf("a==b %t\n", a == b)
		fmt.Printf("a==c %t\n", a == c)
	} else {
		fmt.Printf("not comparable\n")
	}
}

func doStack() {
	list := goutils.ListGoRoutines(true)
	for _, l := range list {
		fmt.Printf("%3d: [%s] %s\n", l.Id, l.Status, l.Current.Name)
		fmt.Printf("     cur:    %s\n", l.Current.Location)
		fmt.Printf("     first:  %s\n", l.First.Name)
		fmt.Printf("     creat:  %s\n", l.Creator.Name)

		for _, f := range l.Stack {
			fmt.Printf("     stack:  %s\n", f.Location)
		}
	}
}

/////////////////////////
func PlainMain() {
	doX()
	doA()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		doStack()
		wg.Done()
	}()
	wg.Wait()
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
