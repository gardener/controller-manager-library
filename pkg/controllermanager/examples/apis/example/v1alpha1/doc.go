// SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// +k8s:deepcopy-gen=package
// +k8s:conversion-gen=github.com/gardener/controller-manager-library/pkg/controllermanager/examples/apis/example
// +k8s:openapi-gen=true
// +k8s:defaulter-gen=TypeMeta

//go:generate gen-crd-api-reference-docs -api-dir . -config ../../../hack/api-reference/api.json -template-dir ../../../hack/api-reference/template -out-file ../../../hack/api-reference/api-v1alpha1.md

// Package v1alpha1 contains example API resources.
// +groupName=example.examples.gardener.cloud
package v1alpha1 // import "github.com/gardener/controller-manager-library/pkg/controllermanager/examples/apis/example/v1alpha1"
