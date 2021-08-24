/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package webhook

import (
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/gardener/controller-manager-library/pkg/config"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/extension"
	"github.com/gardener/controller-manager-library/pkg/resources"
)

type webhook struct {
	extension.ElementBase

	config     *WebhookConfig
	kindconfig config.OptionSource
	extension  *Extension
	definition Definition
	scheme     *runtime.Scheme
	cluster    cluster.Interface
}

var _ Interface = &webhook{}

func NewWebhook(ext *Extension, def Definition, cluster cluster.Interface) (*webhook, error) {
	var err error

	src := def.SchemeSource()
	options := ext.GetConfig().GetSource(def.Name()).(*WebhookConfig)
	if src != nil && cluster != nil {
		cluster, err = ext.GetClusters().Cache().WithScheme(cluster, src)
		if err != nil {
			return nil, err
		}
	}

	var scheme *runtime.Scheme
	if src == nil {
		if cluster != nil {
			scheme = cluster.ResourceContext().Scheme()
		}
	} else {
		scheme = src.Scheme(nil)
	}
	if scheme == nil {
		scheme = resources.DefaultSchemeSource().Scheme(nil)
	}
	if scheme == nil {
		scheme = resources.DefaultScheme()
	}
	this := &webhook{
		extension:  ext,
		definition: def,
		config:     options,
		kindconfig: ext.regctxs[def.Kind()].Config(),
		cluster:    cluster,
		scheme:     scheme,
	}
	this.ElementBase = extension.NewElementBase(ext.GetContext(), ctx_webhook, this, def.Name(), WEBHOOK_SET_PREFIX, options)
	if err != nil {
		return nil, err
	}
	return this, nil
}

func (this *webhook) GetResources() resources.Resources {
	if this.cluster == nil {
		return nil
	}
	return this.cluster.Resources()
}

func (this *webhook) GetEnvironment() Environment {
	return this.extension
}

func (this *webhook) GetKindConfig() config.OptionSource {
	return this.kindconfig
}

func (this *webhook) GetKind() WebhookKind {
	return this.definition.Kind()
}

func (this *webhook) GetDefinition() Definition {
	return this.definition
}

func (this *webhook) GetCluster() cluster.Interface {
	return this.cluster
}

func (this *webhook) GetScheme() *runtime.Scheme {
	return this.scheme
}
