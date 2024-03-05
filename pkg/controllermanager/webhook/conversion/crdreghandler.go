/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package conversion

import (
	"fmt"

	adminreg "k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gardener/controller-manager-library/pkg/config"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/extension"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/webhook"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"github.com/gardener/controller-manager-library/pkg/resources/apiextensions"
)

func init() {
	webhook.RegisterRegistrationHandler(newCRDHandler())
}

type Config struct {
	OmitStorageMigration bool
}

var _ config.OptionSource = (*Config)(nil)

func (this *Config) AddOptionsToSet(set config.OptionSet) {
	set.AddBoolOption(&this.OmitStorageMigration, "omit-crd-storage-migration", "", false, "omit auto migration of crds on changed storage version")
}

////////////////////////////////////////////////////////////////////////////////

type CRDDeclaration struct {
	resources.ObjectData
}

var _ webhook.WebhookDeclaration = (*CRDDeclaration)(nil)

func (this *CRDDeclaration) Kind() webhook.WebhookKind {
	return webhook.CONVERTING
}

////////////////////////////////////////////////////////////////////////////////

type crdhandler struct {
	*webhook.RegistrationHandlerBase
}

func newCRDHandler() *crdhandler {
	return &crdhandler{
		webhook.NewRegistrationHandlerBase(webhook.CONVERTING, &v1beta1.CustomResourceDefinition{}),
	}
}

var _ webhook.RegistrationHandler = (*crdhandler)(nil)

func (this *crdhandler) OptionSourceCreator() extension.OptionSourceCreator {
	return extension.OptionSourceCreatorByExample(&Config{})
}

func (this *crdhandler) RequireDedicatedRegistrations() bool {
	return true
}

func (this *crdhandler) RegistrationNames(def webhook.Definition) []string {
	names := []string{}
	for _, r := range def.Resources() {
		crd := apiextensions.GetCRDs(r.GroupKind())
		if crd == nil {
			continue
		}
		name := crd.Name()
		if name != "" {
			names = append(names, name)
		}
	}
	return names
}

func (this *crdhandler) CreateDeclarations(log logger.LogContext, def webhook.Definition, target cluster.Interface, _ apiextensions.WebhookClientConfigSource) (webhook.WebhookDeclarations, error) {
	result := webhook.WebhookDeclarations{}
	log.Infof("creating crd manifests of %s(%s) for cluster %s(%s)", def.Name(), def.Kind(), target.GetId(), target.GetServerVersion())
	for _, r := range def.Resources() {
		crd := apiextensions.GetCRDFor(r.GroupKind(), target)
		if crd == nil {
			log.Infof("  no crd for %s and cluster version %s", r.GroupKind(), target.GetServerVersion())
			continue
		}
		log.Infof("  %s", crd.GetName())
		result = append(result, &CRDDeclaration{crd})
	}
	log.Infof("done")
	return result, nil
}

func (this *crdhandler) Register(ctx webhook.RegistrationContext, labels map[string]string, cluster cluster.Interface, _ string, declarations ...webhook.WebhookDeclaration) error {
	ctx.Infof("registering crds...")
	log := ctx.AddIndent("  ")
	sub := log.AddIndent("  ")
	for _, d := range declarations {
		decl := d.(*CRDDeclaration).ObjectData
		decl.SetLabels(labels)
		log.Infof("%s (%T)", decl.GetName(), decl)
		err := apiextensions.CreateCRDFromObject(sub, cluster, decl, ctx.Maintainer())
		if err != nil {
			return err
		}
		if ctx.Config().(*Config).OmitStorageMigration {
			sub.Infof("omitting migration check")
		} else {
			sub.Infof("checking for required storage migration")
			err = apiextensions.Migrate(sub, cluster, decl.GetName(), nil)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (this *crdhandler) Delete(log logger.LogContext, name string, def webhook.Definition, cluster cluster.Interface) error {
	resc, err := cluster.Resources().Get(&adminreg.MutatingWebhookConfiguration{})
	if err != nil {
		return err
	}
	if def == nil {
		err := resc.DeleteByName(resources.NewObjectName(name))
		if err != nil && !errors.IsNotFound(err) {
			return fmt.Errorf("error deleting crd %s: %s", name, err)
		}
		return nil
	}

	log.Infof("deleting crds...")
	for _, r := range def.Resources() {
		crd := apiextensions.GetCRDFor(r.GroupKind(), cluster)
		if crd == nil {
			log.Infof("  no crd for %s and cluster version %s", r.GroupKind(), cluster.GetServerVersion())
			continue
		}
		cust, err := cluster.Resources().Get(r)
		if err != nil {
			log.Infof("  resource %s not found: %s", r, err)
			continue
		}
		list, err := cust.List(metav1.ListOptions{})
		if err != nil {
			log.Infof("  list failed for %s: %s", r, err)
			return err
		}
		if len(list) > 0 {
			log.Infof("  deletion os %s skipped because there are still %d objects", r, len(list))
		}
		log.Infof("  %s", crd.GetName())
		err = resources.FilterObjectDeletionError(resc.DeleteByName(resources.NewObjectName(crd.GetName())))
		if err != nil {
			return fmt.Errorf("error deleting crd %s: %s", crd.GetName(), err)
		}
	}
	return nil
}
