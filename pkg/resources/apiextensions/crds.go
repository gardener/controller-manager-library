/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package apiextensions

import (
	"fmt"
	"reflect"
	"time"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/extension"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"github.com/gardener/controller-manager-library/pkg/resources/errors"
	"github.com/gardener/controller-manager-library/pkg/utils"
)

const A_MAINTAINER = "crds.gardener.cloud/maintainer"

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

func (this *CustomResourceDefinition) ObjectFor(cluster resources.Cluster, cp WebhookClientConfigProvider) (resources.Object, error) {
	return cluster.Resources().Wrap(this.DataFor(cluster, cp))
}

func (this *CustomResourceDefinition) DataFor(cluster resources.Cluster, cp WebhookClientConfigProvider) resources.ObjectData {
	if this == nil {
		return nil
	}
	if cp == nil {
		cp = registry
	}
	crd := this.DeepCopy()
	if len(crd.Spec.Versions) > 1 && cp != nil {
		if crd.Spec.Conversion == nil || crd.Spec.Conversion.WebhookClientConfig == nil {
			cfg := cp.GetClientConfig(crd.CRDGroupKind(), cluster)
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
	if cluster.GetServerVersion().LessThan(v122) && (cluster.GetServerVersion().LessThan(v116) || len(crd.Spec.Versions) == 0 || crd.Spec.Versions[0].Schema == nil) {
		o, err := crd.ConvertTo(string(CRD_V1BETA1))
		utils.Must(err)
		return o
	}
	o, err := crd.ConvertTo(string(CRD_V1))
	utils.Must(err)
	return o
}

////////////////////////////////////////////////////////////////////////////////

func CreateCRDFromObject(log logger.LogContext, cluster resources.Cluster, crd resources.ObjectData, maintainer extension.MaintainerInfo) error {
	var err error

	if abs, ok := crd.(*CustomResourceDefinition); ok {
		crd = abs.DataFor(cluster, registry)
	}
	if crd == nil {
		return errors.New(errors.ERR_INVALID, "invalid crd")
	}
	msg := logger.NewOptionalSingletonMessage(log.Infof, "foreign %s", crd.GetName())
	if maintainer.Ident != "" {
		resources.SetAnnotation(crd, A_MAINTAINER, maintainer.Ident)
	}
	found, err := cluster.Resources().GetObject(crd)
	if err == nil {
		if maintainer.ForceCRDUpdate || maintainer.Idents.Contains(found.GetAnnotation(A_MAINTAINER)) {
			msg.ResetWith("uptodate %s", crd.GetName())
			new, _ := resources.GetObjectSpec(crd)
			_, err := found.Modify(func(data resources.ObjectData) (bool, error) {
				mod := false
				if wantsShootNoCleanupLabel(cluster) {
					mod = resources.SetLabel(data, v1beta1constants.ShootNoCleanup, "true")
				}
				spec, _ := resources.GetObjectSpec(data)
				if !reflect.DeepEqual(spec, new) {
					msg.Default("updating %s", crd.GetName())
					if err := resources.SetObjectSpec(data, new); err != nil {
						return true, err
					}
					mod = true
				}
				if v, _ := resources.GetAnnotation(data, A_MAINTAINER); v != maintainer.Ident {
					if maintainer.Ident == "" {
						mod = resources.RemoveAnnotation(data, A_MAINTAINER) || mod
					} else {
						mod = resources.SetAnnotation(data, A_MAINTAINER, maintainer.Ident) || mod
					}
				}
				return mod, nil
			})
			if err != nil {
				log.Errorf("cannot update crd: %s", err)
			}
		}
	} else {
		if errors.IsKind(errors.ERR_UNKNOWN_RESOURCE, err) {
			return err
		}
		if wantsShootNoCleanupLabel(cluster) {
			resources.SetLabel(crd, v1beta1constants.ShootNoCleanup, "true")
		}
		msg.Default("creating %s", crd.GetName())
		err = _CreateCRDFromObject(cluster, crd)
	}
	if err != nil {
		return fmt.Errorf("creating CRD for %s failed: %s", crd.GetName(), err)
	}
	msg.Once()
	return nil
}

func _CreateCRDFromObject(cluster resources.Cluster, crd resources.ObjectData) error {
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
		var versioned resources.ObjectData
		if cluster.GetServerVersion().LessThan(v122) {
			versioned = &apiextensionsv1beta1.CustomResourceDefinition{}
		} else {
			versioned = &apiextensionsv1.CustomResourceDefinition{}
		}
		_, err := cluster.Resources().GetObjectInto(resources.NewObjectName(crdName), versioned)
		if err != nil {
			return false, err
		}
		crd := &apiextensions.CustomResourceDefinition{}
		err = cluster.Resources().Scheme().Convert(versioned, crd, nil)
		if err != nil {
			return false, err
		}

		for _, cond := range crd.Status.Conditions {
			switch cond.Type {
			case apiextensions.Established:
				if cond.Status == apiextensions.ConditionTrue {
					return true, nil
				}
			case apiextensions.NamesAccepted:
				if cond.Status == apiextensions.ConditionFalse {
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

func Migrate(log logger.LogContext, cluster resources.Cluster, crdName string, _ *runtime.Scheme) error {
	key := NewKey(crdName)
	obj, err := cluster.Resources().GetObject(key)
	if err != nil {
		return err
	}
	crd := &apiextensions.CustomResourceDefinition{}
	err = scheme.Convert(obj.Data(), crd, nil)
	if err != nil {
		return err
	}
	stored := ""
	for _, v := range crd.Spec.Versions {
		if v.Storage {
			stored = v.Name
		}
	}
	if stored == "" {
		log.Infof("no stored version indicated for %s: found %s", crdName, utils.Strings(crd.Status.StoredVersions...))
		return nil
	}

	switch len(crd.Status.StoredVersions) {
	case 0:
		log.Infof("no stored versions for %s found (should be %s)", crdName, stored)
		return nil
	case 1:
		if stored == crd.Status.StoredVersions[0] {
			log.Infof("stored versions for %s: %s", crdName, stored)
			return nil
		}
		log.Infof("stored versions mismatch for %s:  found %s (should be %s)", crdName, crd.Status.StoredVersions[0], stored)
		return nil
	default:
		log.Infof("stored versions for %s: %s, required %s", crdName, crd.Status.StoredVersions, stored)
	}

	log.Infof("migration required...")
	gk := resources.NewGroupKind(crd.Spec.Group, crd.Spec.Names.Kind)
	resc, err := cluster.Resources().GetByGK(gk)
	if err != nil {
		return err
	}
	list, err := resc.List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	log.Infof("migrating %d objects", len(list))
	for _, obj := range list {
		nerr := obj.Update()
		if nerr != nil {
			err = nerr
		}
	}
	if err != nil {
		return err
	}
	crd.Status.StoredVersions = []string{stored}
	err = scheme.Convert(crd, obj.Data(), nil)
	if err != nil {
		return err
	}
	return obj.UpdateStatus()
}

func wantsShootNoCleanupLabel(cl resources.Cluster) bool {
	v, _ := cl.GetAttr(cluster.SUBOPTION_CRDS_SHOOT_NO_CLEANUP_LABEL).(bool)
	return v
}
