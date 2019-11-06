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

package main

import (
	"context"
	"fmt"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"log"
)

func main() {
	def := cluster.Configure("default", "kubeconfig", "dummy").Definition()
	c, err := cluster.CreateCluster(context.Background(), logger.New(), def, "default", "")

	if err != nil {
		log.Fatal(err)
	}

	gvk := schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Secret"}
	info, err := c.ResourceContext().Get(gvk)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("SECRET: %s\n", info)
	config := c.Config()
	client, err := dynamic.NewForConfig(&config)
	if err != nil {
		log.Fatal(err)
	}
	rri := client.Resource(info.GroupVersionResource())
	ri := rri.Namespace("default")
	s, err := ri.Get("access", metav1.GetOptions{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("found: %v\n", s)
	u := &unstructured.Unstructured{}

	st, err := c.Resources().Get(gvk)
	if err != nil {
		log.Fatal(err)
	}
	_, err = st.GetInto(resources.NewObjectName("default", "access"), u)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("UNSTRUCTURED RESULT: %t,  %v\n", st.IsUnstructured(), u)

	r, err := c.Resources().GetUnstructuredByGVK(gvk)
	if err != nil {
		log.Fatal(err)
	}
	_, err = r.GetInto(resources.NewObjectName("default", "access"), u)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("UNSTRUCTURED: %t, %v\n", r.IsUnstructured(), u)
	list, err := r.List(metav1.ListOptions{})
	List("UNSTRUCTURED", list, err)

	list, err = r.Namespace("default").ListCached(nil)
	List("UNSTRUCTURED CACHED", list, err)

	list, err = st.Namespace("default").ListCached(nil)
	List("STRUCTURED CACHED", list, err)
}

func List(msg string, list []resources.Object, err error) {
	fmt.Printf("%s:\n", msg)
	if err != nil {
		log.Fatal(err)
	}
	for _, e := range list {
		fmt.Printf("  %s: %T\n", e.ObjectName(), e.Data())
	}
}
