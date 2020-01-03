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

package webhook

import (
	"github.com/gardener/controller-manager-library/pkg/controllermanager"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/webhook/admission"
	areacfg "github.com/gardener/controller-manager-library/pkg/controllermanager/webhook/config"
	adminreg "k8s.io/api/admissionregistration/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

////////////////////////////////////////////////////////////////////////////////
// Definition
////////////////////////////////////////////////////////////////////////////////
const MAIN_CLUSTER = "<main>"

const MUTATING = "mutating"
const VALIDATION = "validating"

type AdmissionHandlerType func(Interface) (admission.Interface, error)

type Environment interface {
	controllermanager.Environment
	GetConfig() *areacfg.Config
}

type Interface interface {
	controllermanager.ElementBase
	admission.Interface

	GetEnvironment() Environment
	GetDefinition() Definition
	GetCluster() cluster.Interface
	GetScheme() *runtime.Scheme
	GetDecoder() *admission.Decoder
}

type OptionDefinition controllermanager.OptionDefinition

type Definition interface {
	GetName() string
	GetResources() []controllermanager.ResourceKey
	GetCluster() string
	GetKind() string
	GetOperations() []adminreg.OperationType
	GetFailurePolicy() adminreg.FailurePolicyType
	GetHandlerType() AdmissionHandlerType
	GetNamespaces() *meta.LabelSelector
	ActivateExplicitly() bool

	ConfigOptions() map[string]OptionDefinition

	Definition() Definition
}
