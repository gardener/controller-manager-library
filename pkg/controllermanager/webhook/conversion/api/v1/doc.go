/*
SPDX-FileCopyrightText: 2019 The Kubernetes Authors.

SPDX-License-Identifier: Apache-2.0
*/

// +k8s:deepcopy-gen=package
// +k8s:protobuf-gen=package
// +k8s:conversion-gen=github.com/gardener/controller-manager-library/pkg/controllermanager/webhook/conversion/api
// +k8s:defaulter-gen=TypeMeta
// +k8s:openapi-gen=true
// +groupName=apiextensions.k8s.io

// Package v1 is the v1 version of the Conversion Review API.
package v1
