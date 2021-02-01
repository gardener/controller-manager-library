// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package informer

import (
	"context"
	"fmt"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	"github.com/gardener/controller-manager-library/pkg/ctxutil"
	"github.com/gardener/controller-manager-library/pkg/goutils"
	"github.com/gardener/controller-manager-library/pkg/logger"
)

type Factory interface {
	NewRawSharedIndexInformer(ctx context.Context, gvk schema.GroupVersionKind) (cache.SharedIndexInformer, error)
}

func Watch(ctx context.Context, factory Factory) {
	ictx:=ctxutil.CancelContext(ctx)
	informer, err := factory.NewRawSharedIndexInformer(ictx,schema.GroupVersionKind{"","v1","ConfigMap"})
	if err != nil {
		fmt.Fprintf(os.Stderr,"failed to create informer: %s\n", err)
		os.Exit(2)
	}
	fmt.Printf("# add handler\n")
	handler := cache.ResourceEventHandlerFuncs{
		AddFunc:    add,
		UpdateFunc: update,
		DeleteFunc: delete,
	}
	informer.AddEventHandler(handler)
	fmt.Printf("# watch\n")
	go informer.Run(ictx.Done())
	if ok := cache.WaitForCacheSync(ictx.Done(), informer.HasSynced); !ok {
		fmt.Fprintf(os.Stderr,"failed to wait for caches to sync")
		os.Exit(2)
	}

	time.Sleep(10 * time.Second)
	fmt.Printf("# stop watch\n")
	ctxutil.Cancel(ictx)
	time.Sleep(10 * time.Second)
}

func InformerMain() {
	ctx := ctxutil.CancelContext(context.Background())

	cluster.Configure("main", "", "").Definition()
	c, err := cluster.CreateCluster(ctx, logger.New(), cluster.Configure("main", "", "").Definition(), "", "")
	if err != nil {
		fmt.Fprintf(os.Stderr,"failed to create kube configmain %s: %s\n", os.Args[1], err)
		os.Exit(2)
	}

	fmt.Printf("# create context\n")
	rctx := c.ResourceContext()
    factory := rctx.(Factory)

	last:=goutils.ListGoRoutines(true)
	count:=0
	for len(last) < 10 {
		count++
		fmt.Printf("\n*** RUN %d (%d goroutines)\n", count, goutils.NumberOfGoRoutines())
    	Watch(ctx,factory)
    	cur:=goutils.ListGoRoutines(true)
    	add, del:= goutils.GoRoutineDiff(last, cur)

    	if len(add) > 0 {
    		fmt.Printf("%s\n", add.ToString("leftover", true))
		}
		if len(del) > 0 {
			fmt.Printf("%s\n", add.ToString("vanished", true))
		}
    	last=cur
	}
}

func add(obj interface{}) {
	//fmt.Printf("add %s\n", obj.(*unstructured.Unstructured).GetName())
	fmt.Printf("A")
}
func update(old, new interface{}) {
	//fmt.Printf("update %s\n", new.(*unstructured.Unstructured).GetName())
	fmt.Printf("U")
}
func delete(obj interface{}) {
	//fmt.Printf("delete %s\n", obj.(*unstructured.Unstructured).GetName())
	fmt.Printf("D")
}
