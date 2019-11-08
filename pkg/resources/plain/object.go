/*
 * Copyright 2019 SAP SE or an SAP affiliate company. All rights reserved.
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

package plain

import (
	"github.com/gardener/controller-manager-library/pkg/resources/abstract"
)

type _object struct {
	*abstract.AbstractObject
}

var _ Object = &_object{}

func NewObject(data ObjectData, resource Interface) Object {
	return &_object{abstract.NewAbstractObject(data, resource)}
}

func (this *_object) DeepCopy() Object {
	data := this.ObjectData.DeepCopyObject().(ObjectData)
	return NewObject(data, this.GetResource())
}

/////////////////////////////////////////////////////////////////////////////////

func (this *_object) GetResource() Interface {
	return this.AbstractObject.GetResource().(Interface)
}

func (this *_object) Resources() Resources {
	return this.AbstractObject.GetResource().(Interface).Resources()
}
