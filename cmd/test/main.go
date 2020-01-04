package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gardener/controller-manager-library/cmd/test/certs"
	"github.com/gardener/controller-manager-library/cmd/test/cond"
	"github.com/gardener/controller-manager-library/cmd/test/config"
	"github.com/gardener/controller-manager-library/cmd/test/errors"
	"github.com/gardener/controller-manager-library/cmd/test/field"
	"github.com/gardener/controller-manager-library/cmd/test/match"
	"github.com/gardener/controller-manager-library/cmd/test/plain"
	"github.com/gardener/controller-manager-library/cmd/test/preferred"
	"github.com/gardener/controller-manager-library/cmd/test/scheme"
	"github.com/gardener/controller-manager-library/pkg/sync"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller"

	_ "github.com/gardener/controller-manager-library/pkg/resources/defaultscheme"
)

var values = map[controller.ResourceKey]int{}

func main() {

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
		case "configmain":
			config.ConfigMain()
		case "errors":
			errors.ErrorsMain()
		case "cond":
			cond.CondMain()
		case "plain":
			plain.PlainMain()
		case "preferred":
			preferred.PreferredMain()
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

/*
////////////////////////////////////////////////////////////////////////////////

type R interface {
	Name() string
}

type Getter interface {
	Get(interface{}) R
}


type MyR interface {
	R
	Other()
}

type MyGetter interface {
	Get(interface{}) MyR
}

func X(g Getter) {

}

func DO() {
	var G MyGetter

	X(G)
}
*/
