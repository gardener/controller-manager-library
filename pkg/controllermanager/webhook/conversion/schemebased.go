/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package conversion

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/webhook"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
)

func validateScheme(wh webhook.Interface) error {
	scheme := wh.GetScheme()
outer:
	for _, r := range wh.GetDefinition().Resources() {
		gk := r.GroupKind()
		if scheme.IsGroupRegistered(gk.Group) {
			for gvk := range scheme.AllKnownTypes() {
				if gvk.GroupKind() == gk {
					continue outer
				}
			}
			return fmt.Errorf("scheme does not contain any version for %s", gk)
		} else {
			return fmt.Errorf("scheme does not contain group %s", gk.Group)
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// type
////////////////////////////////////////////////////////////////////////////////

type schemehandler struct {
	scheme  *runtime.Scheme
	decoder *resources.Decoder
}

func schemetype(wh webhook.Interface) (Interface, error) {
	return &schemehandler{wh.GetScheme(), resources.NewDecoder(wh.GetScheme())}, nil
}

var _ ConversionHandlerType = schemetype

func (this *schemehandler) Handle(log logger.LogContext, version string, obj runtime.RawExtension) (runtime.Object, error) {
	versions := resources.VersionedObjects{}
	err := this.decoder.DecodeInto(obj.Raw, &versions)
	if err != nil {
		return nil, err
	}
	first := versions.First().(resources.ObjectData)
	gvk := first.GetObjectKind().GroupVersionKind()
	gv, err := schema.ParseGroupVersion(version)
	if err != nil {
		return nil, err
	}
	log.Infof("  converting %s/%s(%s)", first.GetName(), first.GetNamespace(), gvk)
	gvk.Version = gv.Version
	gvk.Group = gv.Group
	conv, err := this.scheme.New(gvk)
	if err != nil {
		return nil, err
	}
	err = this.scheme.Convert(versions.Last(), conv, nil)
	if err != nil {
		return nil, err
	}
	conv.GetObjectKind().SetGroupVersionKind(gvk)
	return conv, nil
}

func SchemeBasedConversion() *configuration {
	return &configuration{_Definition{factory: schemetype, validator: webhook.ValidatorFunc(validateScheme)}}
}
