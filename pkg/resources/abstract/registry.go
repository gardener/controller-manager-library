/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package abstract

import (
	"errors"
	"fmt"
	"sync"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var lock sync.Mutex

///////////////////////////////////////////////////////////////////////////////
// Explcit version mappings for api groups to use for resources

var defaultVersions = map[string]string{}

func DeclareDefaultVersion(gv schema.GroupVersion) {
	lock.Lock()
	defer lock.Unlock()

	if old, ok := defaultVersions[gv.Group]; ok {
		panic(fmt.Sprintf("default version for %s already set to %s", gv, old))
	}
	defaultVersions[gv.Group] = gv.Version
}

func DefaultVersion(g string) string {
	lock.Lock()
	defer lock.Unlock()
	return defaultVersions[g]
}

///////////////////////////////////////////////////////////////////////////////
// registration of default schemes for info management

var scheme = runtime.NewScheme()

func Register(builders ...runtime.SchemeBuilder) error {
	lock.Lock()
	defer lock.Unlock()
	var errs []error
	for _, b := range builders {
		if err := b.AddToScheme(scheme); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func DefaultScheme() *runtime.Scheme {
	return scheme
}
