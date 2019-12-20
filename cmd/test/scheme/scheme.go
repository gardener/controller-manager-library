package scheme

import (
	"context"
	"fmt"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	"github.com/gardener/controller-manager-library/pkg/ctxutil"
	"github.com/gardener/controller-manager-library/pkg/kutil"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"os"
	"time"
)

var listGroups = false
var listScheme = false

type REQ interface {
	Visible()
}
type IF interface {
}

type ST struct {
	IF
}

type Test struct {
}

func (t *Test) Visible() {

}

func SchemeMain() {

	{
		s := &ST{&Test{}}
		var i interface{} = s
		_, ok := i.(REQ)
		fmt.Printf("indirectly visible %t\n", ok)
		_, ok = s.IF.(REQ)
		fmt.Printf("directly   visible %t\n", ok)
	}
	if len(os.Args) <= 1 {
		fmt.Printf("kubeconfig missing\n")
		os.Exit(1)
	}

	ctx := ctxutil.CancelContext(context.Background())

	cluster.Configure("main", "", "").Definition()
	c, err := cluster.CreateCluster(ctx, logger.New(), cluster.Configure("main", "", "").Definition(), "", os.Args[1])
	if err != nil {
		fmt.Errorf("failed to create kube configmain %p: %s", os.Args[1], err)
		os.Exit(2)
	}

	/*
		a, err := c.ServerResources()
		if err != nil {
			fmt.Errorf("failed to get group version resources: %s", err)
			os.Exit(1)
		}
		for _, r := range a {
			gv, _ := schema.ParseGroupVersion(r.GroupVersion)
			fmt.Printf("***************************** %s\n", gv)
			for _, i := range r.APIResources {
				fmt.Printf("%#v\n", i)
				//fmt.Printf("%s (%s) %s %s %t\n", i.Group, i.Version, i.Name, i.GroupKind, i.Namespaced)
			}
		}
	*/

	s := runtime.NewScheme()
	corev1.AddToScheme(s)
	if listScheme {
		for gvk, t := range s.AllKnownTypes() {
			l := kutil.DetermineListType(s, gvk.GroupVersion(), t)
			fmt.Printf("%s: %s [%s]\n", gvk, t, l)
		}
	}

	fmt.Printf("# create context\n")
	rctx, err := resources.NewResourceContext(ctx, c, s, 30*time.Second)
	if err != nil {
		fmt.Printf("cannot create resource context: %s\n", err)
		os.Exit(3)
	}

	fmt.Printf("# get factory\n")
	factory := rctx.SharedInformerFactory()

	if listGroups {
		for _, gv := range rctx.GetGroups() {
			fmt.Printf("**** %s ****\n", gv)
			for _, r := range rctx.GetResourceInfos(gv) {
				fmt.Printf("  %s\n", r.InfoString())
			}
		}
	}

	fmt.Printf("# get informer\n")
	informer, err := factory.InformerForObject(&corev1.ConfigMap{})
	if err != nil {
		fmt.Printf("cannot create informer: %s\n", err)
		os.Exit(4)
	}
	informer, err = factory.InformerForObject(&corev1.ConfigMap{})
	informer, err = factory.InformerForObject(&corev1.ConfigMap{})

	fmt.Printf("# get generic uncached configamp\n")

	h, err := rctx.Resources().Get(schema.GroupKind{"", "ConfigMap"})
	if err != nil {
		fmt.Printf("cannot get uncached resource: %s\n", err)
		os.Exit(5)
	}

	fmt.Printf("# add handler\n")
	h.AddEventHandler(resources.ResourceEventHandlerFuncs{
		AddFunc:    add,
		UpdateFunc: update,
		DeleteFunc: delete,
	})
	factory.Start(ctx.Done())
	factory.WaitForCacheSync(ctx.Done())

	fmt.Printf("# get cached from informer\n")
	data, err := informer.Lister().Namespace("default").Get("test")
	if err != nil {
		fmt.Printf("cannot get object: %s\n", err)
	} else {
		fmt.Printf("* GOT %#v\n", data)
	}

	fmt.Printf("# get cached from resource\n")
	obj, err := h.Namespace("default").GetCached("test")
	if err != nil {
		fmt.Printf("cannot get cached object: %s\n", err)
	} else {
		fmt.Printf("** GOT %#v\n", obj.Data())
	}

	fmt.Printf("# get uncached from resource\n")
	obj, err = h.Namespace("default").Get("test")
	if err != nil {
		fmt.Printf("cannot get uncached object: %s\n", err)
	} else {
		fmt.Printf("*** GOT %#v\n", obj.Data())
	}

	fmt.Printf("# get wrapper\n")
	r, err := rctx.Resources().GetObject(data)
	if err != nil {
		fmt.Printf("cannot get object: %s\n", err)
		os.Exit(6)
	}
	fmt.Printf("**** GOT %#v\n", r.Data())

	fmt.Printf("# set finalizer\n")
	err = r.SetFinalizer("test/mandelsoft.org")
	if err != nil {
		fmt.Printf("cannot set finalizer: %s\n", err)
		os.Exit(7)
	}

	fmt.Printf("# list uncached\n")
	l, err := h.List(metav1.ListOptions{})
	if err != nil {
		fmt.Printf("cannot get uncached list: %s\n", err)
		os.Exit(8)
	}
	if l != nil {
		for i, o := range l {
			fmt.Printf("List %d: %#v", i, o.Data())
		}
	}
}

func add(obj resources.Object) {
	fmt.Printf("add %s\n", obj.Data())
}
func update(old, new resources.Object) {
	fmt.Printf("update %s\n", new.Data())
}
func delete(obj resources.Object) {
	fmt.Printf("delete %s\n", obj.Data())
}
