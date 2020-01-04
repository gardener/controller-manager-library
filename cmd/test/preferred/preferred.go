/*
 * Copyright 2020 SAP SE or an SAP affiliate company. All rights reserved.
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

package preferred

import (
	"context"
	"fmt"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"os"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	"github.com/gardener/controller-manager-library/pkg/ctxutil"
	"github.com/gardener/controller-manager-library/pkg/kutil"
	"github.com/gardener/controller-manager-library/pkg/logger"

	adminregv1 "k8s.io/api/admissionregistration/v1"
	adminregv1beta1 "k8s.io/api/admissionregistration/v1beta1"

	"k8s.io/apimachinery/pkg/runtime"
)

func PreferredMain() {

	ctx := ctxutil.CancelContext(context.Background())
	scheme := runtime.NewScheme()
	adminregv1.AddToScheme(scheme)
	adminregv1beta1.AddToScheme(scheme)

	fmt.Printf("create cluster\n")
	def := cluster.Configure("main", "", "").Scheme(scheme).Definition()
	c, err := cluster.CreateCluster(ctx, logger.New(), def, "", "")
	if err != nil {
		fmt.Errorf("failed to create cluster: %s", err)
		os.Exit(2)
	}
	fmt.Printf("got cluster\n")

	for gvk, t := range scheme.AllKnownTypes() {
		l := kutil.DetermineListType(scheme, gvk.GroupVersion(), t)
		fmt.Printf("%s: %s [%s]\n", gvk, t, l)
	}
	fmt.Println()

	rctx := c.ResourceContext()
	for _, gv := range rctx.GetGroups() {
		fmt.Printf("**** %s ****\n", gv)
		for _, r := range rctx.GetResourceInfos(gv) {
			fmt.Printf("  %s\n", r.InfoString())
		}
	}

	fmt.Println()

	rs := c.Resources()

	gk := resources.NewGroupKind(adminregv1.GroupName, "ValidatingWebhookConfiguration")
	r, err := rs.Get(gk)
	if err != nil {
		fmt.Errorf("failed to get resource: %s", err)
		os.Exit(2)
	}
	fmt.Printf("%s\n", r.Info())
}
