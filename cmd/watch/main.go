/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package main

import (
	"context"
	"fmt"
	"os"
	"reflect"

	"github.com/gardener/controller-manager-library/pkg/kutil"
	_ "github.com/gardener/controller-manager-library/pkg/resources/defaultscheme/v1.16"
	"github.com/gardener/controller-manager-library/pkg/resources/minimal"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
)

func check(err error) {
	if err != nil {
		fmt.Fprint(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func main() {

	ctx, _ := context.WithCancel(context.Background())
	def := cluster.Configure("default", "kubeconfig", "dummy").Definition()
	c, err := cluster.CreateCluster(context.Background(), logger.New(), def, "default", "")
	check(err)

	res, err := c.Resources().Get(&v1.ConfigMap{})
	check(err)

	info := res.Info()
	client, err := c.ResourceContext().GetClient(res.GroupVersionKind().GroupVersion())
	check(err)

	elemType := reflect.TypeOf(minimal.MinimalObject{})
	//elemType=res.ObjectType()

	listType := kutil.DetermineListType(c.ResourceContext().Scheme(), res.GroupVersionKind().GroupVersion(), elemType)

	informer := create(client, info, elemType, listType, c.ResourceContext().GetParameterCodec())

	informer.AddEventHandler(&handler{})
	informer.Run(ctx.Done())

}

func create(client restclient.Interface, res *resources.Info, elemType reflect.Type, listType reflect.Type, paramcodec runtime.ParameterCodec) cache.SharedIndexInformer {
	logger.Infof("new generic informer for %s (%s) %s", elemType, res.GroupVersionKind(), listType)
	indexers := cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}
	ctx := context.TODO()
	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				result := reflect.New(listType).Interface().(runtime.Object)
				r := client.Get().
					Resource(res.Name()).
					VersionedParams(&options, paramcodec)
				if res.Namespaced() {
					r = r.Namespace("")
				}

				return result, r.Do(ctx).Into(result)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				options.Watch = true
				r := client.Get().
					Resource(res.Name()).
					VersionedParams(&options, paramcodec)
				if res.Namespaced() {
					r = r.Namespace("")
				}

				w, err := r.Watch(ctx)
				if elemType == reflect.TypeOf(minimal.MinimalObject{}) {
					w = minimal.MinimalWatchFilter(w)
				}
				return w, err
			},
		},
		reflect.New(elemType).Interface().(runtime.Object),
		0,
		indexers,
	)
	return informer
}

type handler struct {
}

func (this *handler) toString(obj interface{}) string {
	t := "<none>"
	if o, ok := obj.(runtime.Object); ok {
		t = o.GetObjectKind().GroupVersionKind().String()
	}
	if m, ok := obj.(metav1.Object); ok {
		return fmt.Sprintf("%s/%s (%s)", m.GetNamespace(), m.GetName(), t)
	}
	return fmt.Sprintf("%+v (%s)", obj, t)
}

func (this *handler) OnAdd(obj interface{}) {
	logger.Infof("add %T: %s", obj, this.toString(obj))
}
func (this *handler) OnUpdate(old, new interface{}) {
	logger.Infof("update %T: %s", new, this.toString(new))
}
func (this *handler) OnDelete(obj interface{}) {
	logger.Infof("del %T: %s", obj, this.toString(obj))
}
