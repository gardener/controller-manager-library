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

package apiextensions

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/gardener/controller-manager-library/pkg/resources/errors"
	"github.com/gardener/controller-manager-library/pkg/utils"

	"github.com/gardener/controller-manager-library/pkg/resources"
)

type CRDSpecification interface{}

type CustomResourceDefinition struct {
	*apiextensions.CustomResourceDefinition
}

func (this *CustomResourceDefinition) DeepCopyObject() runtime.Object {
	return this.DeepCopy()
}

func (this *CustomResourceDefinition) DeepCopy() *CustomResourceDefinition {
	return &CustomResourceDefinition{this.CustomResourceDefinition.DeepCopyObject().(*apiextensions.CustomResourceDefinition)}
}

func (this *CustomResourceDefinition) CRDVersions() []string {
	r := []string{}
	for _, v := range this.Spec.Versions {
		r = append(r, v.Name)
	}
	return r
}

func (this *CustomResourceDefinition) CRDGroupKind() schema.GroupKind {
	return resources.NewGroupKind(this.Spec.Group, this.Spec.Names.Kind)
}

func (this *CustomResourceDefinition) ConvertTo(v string) (resources.ObjectData, error) {
	gvk := schema.GroupVersionKind{
		Group:   apiextensions.GroupName,
		Version: v,
		Kind:    "CustomResourceDefinition",
	}

	new, err := scheme.New(gvk)
	if err != nil {
		return nil, err
	}
	err = scheme.Convert(this.CustomResourceDefinition, new, nil)
	if err != nil {
		return nil, err
	}
	new.GetObjectKind().SetGroupVersionKind(gvk)
	return new.(resources.ObjectData), nil
}

func (this *CustomResourceDefinition) CRDRestrict(versions ...string) (*CustomResourceDefinition, error) {
	new := this.DeepCopy()

	vers := []apiextensions.CustomResourceDefinitionVersion{}
outer:
	for _, v := range versions {
		for _, e := range new.Spec.Versions {
			if e.Name == v {
				vers = append(vers, e)
				continue outer
			}
		}
		return nil, errors.ErrUnknown.New(v)
	}
	new.Spec.Versions = vers
	return new, nil
}

func (this *CustomResourceDefinition) For(cluster resources.Cluster) resources.ObjectData {
	if this == nil {
		return nil
	}
	crd := this.DeepCopy()
	if len(crd.Spec.Versions) > 1 {
		fmt.Printf("==========%d\n", len(crd.Spec.Versions))
		if crd.Spec.Conversion == nil || crd.Spec.Conversion.WebhookClientConfig == nil {
			cfg := GetClientConfig(crd.CRDGroupKind(), cluster)
			if cfg != nil {
				if crd.Spec.Conversion == nil || crd.Spec.Conversion.Strategy == apiextensions.NoneConverter {
					crd.Spec.Conversion = &apiextensions.CustomResourceConversion{
						Strategy:                 apiextensions.WebhookConverter,
						ConversionReviewVersions: []string{string(CRD_V1), string(CRD_V1BETA1)},
					}
				}
				crd.Spec.Conversion.WebhookClientConfig = toClientConfig(cfg.WebhookClientConfig())
			} else {
				fmt.Printf("========== no client config\n")
			}
		}
	}
	if cluster.GetServerVersion().LessThan(v116) {
		o, err := crd.ConvertTo(string(CRD_V1BETA1))
		utils.Must(err)
		return o
	}
	o, err := crd.ConvertTo(string(CRD_V1))
	utils.Must(err)
	return o
}

////////////////////////////////////////////////////////////////////////////////

func CreateCRDObjectWithStatus(groupName, version, rkind, rplural, shortName string, namespaces bool, columns ...v1beta1.CustomResourceColumnDefinition) *v1beta1.CustomResourceDefinition {
	return _CreateCRDObject(true, groupName, version, rkind, rplural, shortName, namespaces, columns...)
}

func CreateCRDObject(groupName, version, rkind, rplural, shortName string, namespaces bool, columns ...v1beta1.CustomResourceColumnDefinition) *v1beta1.CustomResourceDefinition {
	return _CreateCRDObject(false, groupName, version, rkind, rplural, shortName, namespaces, columns...)
}

func _CreateCRDObject(status bool, groupName, version, rkind, rplural, shortName string, namespaces bool, columns ...v1beta1.CustomResourceColumnDefinition) *v1beta1.CustomResourceDefinition {
	crdName := rplural + "." + groupName
	scope := v1beta1.ClusterScoped
	if namespaces {
		scope = v1beta1.NamespaceScoped
	}
	crd := &v1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: crdName,
		},
		Spec: v1beta1.CustomResourceDefinitionSpec{
			Group:   groupName,
			Version: version,
			Scope:   scope,
			Names: v1beta1.CustomResourceDefinitionNames{
				Plural: rplural,
				Kind:   rkind,
			},
		},
	}

	if status {
		crd.Spec.Subresources = &v1beta1.CustomResourceSubresources{Status: &v1beta1.CustomResourceSubresourceStatus{}}
	}
	for _, c := range columns {
		crd.Spec.AdditionalPrinterColumns = append(crd.Spec.AdditionalPrinterColumns, c)
	}
	crd.Spec.AdditionalPrinterColumns = append(crd.Spec.AdditionalPrinterColumns, v1beta1.CustomResourceColumnDefinition{Name: "AGE", Type: "date", JSONPath: ".metadata.creationTimestamp"})

	if len(shortName) != 0 {
		crd.Spec.Names.ShortNames = []string{shortName}
	}

	return crd
}

func CreateCRD(cluster resources.Cluster, groupName, version, rkind, rplural, shortName string, namespaces bool, columns ...v1beta1.CustomResourceColumnDefinition) error {
	crd := CreateCRDObject(groupName, version, rkind, rplural, shortName, namespaces, columns...)
	return CreateCRDFromObject(cluster, crd)
}

func CreateCRDFromObject(cluster resources.Cluster, crd resources.ObjectData) error {
	resc, err := cluster.Resources().GetByExample(crd)
	if err != nil {
		return err
	}
	if resc.GroupKind() != crdGK {
		return errors.ErrUnexpectedResource.New("custom resource definition", resc.GroupKind())
	}
	_, err = resc.Create(crd)
	if err != nil && !k8serr.IsAlreadyExists(err) {
		return errors.ErrFailed.Wrap(err, "create CRD", crd.GetName())
	}
	return WaitCRDReady(cluster, crd.GetName())
}

func WaitCRDReady(cluster resources.Cluster, crdName string) error {
	err := wait.PollImmediate(5*time.Second, 60*time.Second, func() (bool, error) {
		crd := &v1beta1.CustomResourceDefinition{}
		_, err := cluster.Resources().GetObjectInto(resources.NewObjectName(crdName), crd)
		if err != nil {
			return false, err
		}
		for _, cond := range crd.Status.Conditions {
			switch cond.Type {
			case v1beta1.Established:
				if cond.Status == v1beta1.ConditionTrue {
					return true, nil
				}
			case v1beta1.NamesAccepted:
				if cond.Status == v1beta1.ConditionFalse {
					return false, errors.New(errors.ERR_CONFLICT,
						"CRD Name conflict for '%s': %v", crdName, cond.Reason)
				}
			}
		}
		return false, nil
	})
	if err != nil {
		return errors.ErrFailed.Wrap(err, "wait for CRD creation", crdName)
	}
	return nil
}
