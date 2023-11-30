/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package reconcilers

import (
	"fmt"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller"
	"github.com/gardener/controller-manager-library/pkg/resources"
)

func ClusterResources(cluster string, gks ...schema.GroupKind) Resources {
	return func(c controller.Interface) ([]resources.Interface, error) {
		result := []resources.Interface{}
		resources := c.GetCluster(cluster).Resources()
		for _, gk := range gks {
			res, err := resources.Get(gk)
			if err != nil {
				return nil, fmt.Errorf("resources type %s not found: %s", gk, err)
			}
			result = append(result, res)
		}
		return result, nil
	}
}

func MainResources(gks ...schema.GroupKind) Resources {
	return ClusterResources("", gks...)
}

func listCachedWithNamespace(r resources.Interface, namespace string) ([]resources.Object, error) {
	if namespace != "" {
		return r.Namespace(namespace).ListCached(labels.Everything())
	}
	return r.ListCached(labels.Everything())
}

func AsKeySet(key resources.ClusterObjectKey) resources.ClusterObjectKeySet {
	if key.Name() == "" {
		return resources.NewClusterObjectKeySet()
	}
	return resources.NewClusterObjectKeySet(key)
}
