/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
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
	c, err := cluster.CreateCluster(ctx, logger.New(), def, "", nil)
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
	fmt.Printf("=============================\n")
	for _, gv := range rctx.GetGroups() {
		fmt.Printf("**** %s ****\n", gv)
		for _, r := range rctx.GetResourceInfos(gv) {
			fmt.Printf("  %s\n", r.InfoString())
		}
	}
	fmt.Printf("=============================\n")

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
