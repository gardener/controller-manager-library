/*
 * Copyright 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
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
