// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/gardener/controller-manager-library/cmd/test/certs"
	"github.com/gardener/controller-manager-library/cmd/test/cond"
	"github.com/gardener/controller-manager-library/cmd/test/config"
	"github.com/gardener/controller-manager-library/cmd/test/errors"
	"github.com/gardener/controller-manager-library/cmd/test/field"
	"github.com/gardener/controller-manager-library/cmd/test/match"
	"github.com/gardener/controller-manager-library/cmd/test/misc"
	"github.com/gardener/controller-manager-library/cmd/test/plain"
	"github.com/gardener/controller-manager-library/cmd/test/preferred"
	"github.com/gardener/controller-manager-library/cmd/test/recover"
	"github.com/gardener/controller-manager-library/cmd/test/scheme"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller"
	"github.com/gardener/controller-manager-library/pkg/sync"
	"github.com/gardener/controller-manager-library/pkg/utils"

	_ "github.com/gardener/controller-manager-library/pkg/controllermanager/examples/apis/example/crds"
	_ "github.com/gardener/controller-manager-library/pkg/resources/defaultscheme/v1.18"
)

var values = map[controller.ResourceKey]int{}

type V struct {
	V interface{}
}

func PV(v interface{}) {
	pv("", "", reflect.ValueOf(v))
}

func pv(gap, prefix string, v reflect.Value) {
	g := gap + "  "
	k := fmt.Sprintf("%s%s%s (%s)", gap, prefix, v.Kind().String(), v.Type())
	if utils.IsNil(v) {
		fmt.Printf("%s <NIL>\n", k)
		return
	}
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		fmt.Printf("%s elem:\n", k)
		pv(g, "", v.Elem())
	case reflect.Struct:
		fmt.Printf("%s fields %d:\n", k, v.NumField())
		for i := 0; i < v.NumField(); i++ {
			f := v.Type().Field(i)
			p := fmt.Sprintf("%s: ", f.Name)
			pv(g, p, v.Field(i))
		}
	case reflect.Map:
		fmt.Printf("%s len %d:\n", k, v.Len())
		for it := v.MapRange(); it.Next(); {
			p := fmt.Sprintf("%s: ", it.Value().String())
			pv(g, p, it.Value())
		}
	case reflect.Array, reflect.Slice:
		fmt.Printf("%s len %d:\n", k, v.Len())
		for i := 0; i < v.Len(); i++ {
			p := fmt.Sprintf("%d: ", i)
			pv(g, p, v.Index(i))
		}
	default:
		fmt.Printf("%s (%s) %q\n", k, v.Type(), v.String())
	}
}

func doStruct() {
	a := []V{
		{"test"},
	}
	fmt.Printf("----------------------\n")
	av := reflect.ValueOf(a)
	PV(a)

	v := av.Index(0)

	fmt.Printf("%s: %s\n", v.Type(), v)
	v.Set(reflect.ValueOf(V{"bla"}))
	fmt.Printf("v: %s\n", a[0].V)
}

type I interface {
	Error() string
}

type E string

func (this E) Error() string {
	return string(this)
}

func doTypedInterface() {
	other := E("other")
	a := []I{
		fmt.Errorf("test"),
		nil,
	}
	fmt.Printf("----------------------\n")
	av := reflect.ValueOf(a)
	av = reflect.Append(av, reflect.ValueOf(other))
	av = reflect.Append(av, reflect.New(av.Type().Elem()).Elem())
	PV(av.Interface())

	v := av.Index(0)

	fmt.Printf("%s: %s\n", v.Type(), v)
	v.Set(reflect.ValueOf(fmt.Errorf("bla")))
	fmt.Printf("v: %s\n", a[0])
}

func doInterface() {
	a := []interface{}{
		"test",
		nil,
	}
	fmt.Printf("----------------------\n")
	av := reflect.ValueOf(a)
	av = reflect.Append(av, reflect.New(av.Type().Elem()).Elem())
	PV(av.Interface())

	v := av.Index(0)

	fmt.Printf("%s: %s\n", v.Type(), v)
	v.Set(reflect.ValueOf("bla"))
	fmt.Printf("v: %s\n", a[0])
}

func main() {

	/*
		PV(E("err"))
		doStruct()
		doInterface()
		doTypedInterface()
		os.Exit(0)

	*/

	//doit()
	for i := 1; i < len(os.Args); i++ {
		fmt.Printf("*** %s ***\n", os.Args[i])
		switch os.Args[i] {
		case "field":
			field.FieldMain()

		case "scheme":
			scheme.SchemeMain()
		case "match":
			match.MatchMain()
		case "certs":
			certs.CertsMain()
		case "config":
			config.ConfigMain()
		case "errors":
			errors.ErrorsMain()
		case "cond":
			cond.CondMain()
		case "plain":
			plain.PlainMain()
		case "preferred":
			preferred.PreferredMain()
		case "misc":
			misc.MiscMain()
		case "recover":
			recover.RecoverMain()
		}
	}
}

type Interface interface {
	Func()
}

type A struct{}
type B struct{}

type Common struct {
	Interface
	*B
}

func (*A) Func() {
	fmt.Printf("A.Func\n")
}
func (*B) Func() {
	fmt.Printf("B.Func\n")
}

func main0() {
	//c := &Common{&A{}, &B{}}
	//c.Func()
}

func main1() {
	k1 := controller.NewResourceKey("a", "b")
	k2 := controller.NewResourceKey("a", "b")
	values[k1] = 1
	fmt.Printf("k1: %d\n", values[k1])
	fmt.Printf("k2: %d\n", values[k2])

	fmt.Printf("cluster mapping: %s", set)
}

type C struct {
	name string
}

type S struct {
	m map[string]*C
}

var set = &S{map[string]*C{"a": {"A"}}}

func (c *C) String() string {
	return c.name
}
func (c *S) String() string {
	return fmt.Sprintf("%v", c.m)[3:]
}

func doit() {
	fmt.Println("sync test *******************")
	s1 := &sync.SyncPoint{}

	ctx, cancel := context.WithCancel(context.TODO())
	go func() {
		time.Sleep(10 * time.Second)
		fmt.Println("reaching sync point")
		s1.Reach()
	}()
	for i := time.Duration(0); i < 5; i++ {
		go func(i time.Duration) {
			time.Sleep(i * 3 * time.Second)
			fmt.Println("check")
			if s1.Sync(ctx) {
				fmt.Println("sync point reached")
			} else {
				fmt.Println("aborted")
			}
		}(i)
	}

	cancel()
	time.Sleep(15 * time.Second)
}
