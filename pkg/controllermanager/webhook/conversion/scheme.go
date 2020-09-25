/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package conversion

import (
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/webhook/conversion/api"
	v1 "github.com/gardener/controller-manager-library/pkg/controllermanager/webhook/conversion/api/v1"
	v1beta1 "github.com/gardener/controller-manager-library/pkg/controllermanager/webhook/conversion/api/v1beta1"
	"github.com/gardener/controller-manager-library/pkg/resources"
)

var reviewScheme = runtime.NewScheme()
var reviewDecoder *resources.Decoder

func init() {
	utilruntime.Must(v1.AddToScheme(reviewScheme))
	utilruntime.Must(v1beta1.AddToScheme(reviewScheme))
	utilruntime.Must(api.AddToScheme(reviewScheme))
	reviewDecoder = resources.NewDecoder(reviewScheme)
}
