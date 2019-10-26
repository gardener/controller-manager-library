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

package errors

import "github.com/gardener/controller-manager-library/pkg/errors"

const (
	GROUP = "gardener/cml/resources"

	// formal

	// ERR_UNEXPECTED_TYPE: general error for an unexpected type
	// objs:
	// - usage scenario
	// - object whose type is unexpected
	ERR_UNEXPECTED_TYPE = "UNEXPECTED_TYPE"

	// ERR_UNKNOWN: general error for an unknown element
	// objs:
	// - unknown element
	ERR_UNKNOWN = "UNKNOWN"

	// ERR_FAILED: operation failed for object
	// objs:
	// - operation
	// - object
	ERR_FAILED = "FAILED"

	// ERR_NAMESPACED: resource is namespaced and requires namespace for identity
	// objs:
	// - element type info, i.e gvk
	ERR_NAMESPACED = "NAMESPACED"

	// ERR_NOT_NAMESPACED: resource is not namespaced
	// objs:
	// - element type info, i.e gvk
	ERR_NOT_NAMESPACED = "NOT_NAMESPACED"

	// ERR_RESOURCE_MISMATCH: resource object cannot handle instance of foreign resource
	// objs:
	// - called resource
	// - requested resource
	ERR_RESOURCE_MISMATCH = "RESOURCE_MISMATCH"

	// ERR_TYPE_MISMATCH: wrong type given
	// objs:
	// - given object
	// - required type
	ERR_TYPE_MISMATCH = "TYPE_MISMATCH"

	// ERR_NO_STATUS_SUBRESOURCE: resource has no status sub resource
	// objs:
	// - given resource spec
	ERR_NO_STATUS_SUBRESOURCE = "NO_STATUS_SUBRESOURCE"

	// informal

	ERR_OBJECT_REJECTED = "OBJECT_REJECTED"
	// objs: key

	ERR_NO_LIST_TYPE = "NO_LIST_TYPE"
	// objs: type with missing list type

	ERR_NON_UNIQUE_MAPPING = "NON_UNIQUE_MAPPING"
	// objs: key

	ERR_INVALID = "INVALID"
	// objs: some invalid element

	ERR_INVALID_RESPONSE = "INVALID_RESPONSE"
	// objs:
	// - source
	// - response

	ERR_PERMISSION_DENIED = "PERMISSION_DENIED"
	// objs:
	// - source key
	// - relation
	// - used key

	ERR_CONFLICT = "CONFLICT"
	// objs:
	// - target
	// - reason

)

var (
	// ErrTypeMismatch
	ErrTypeMismatch = errors.DeclareFormalType(GROUP, ERR_TYPE_MISMATCH, "unexpected type %T (expected %s)")
	// ErrUnexpectedType: invalid type for a dedicated use case
	ErrUnexpectedType = errors.DeclareFormalType(GROUP, ERR_UNEXPECTED_TYPE, "unexpected type for %s: %T")
	// ErrUnknown
	ErrUnknown = errors.DeclareFormalType(GROUP, ERR_UNKNOWN, "unknown %s")
	// ErrUnknown
	ErrFailed = errors.DeclareFormalType(GROUP, ERR_FAILED, "%s failed: %s")
	// ErrNamespaced
	ErrNamespaced = errors.DeclareFormalType(GROUP, ERR_NAMESPACED, "resource is namespaced: %s")
	// ErrNotNamespaced
	ErrNotNamespaced = errors.DeclareFormalType(GROUP, ERR_NOT_NAMESPACED, "resource is not namespaced: %s")
	// ErrResourceMismatch
	ErrResourceMismatch = errors.DeclareFormalType(GROUP, ERR_NAMESPACED, "resource object for %s cannot handle resource %s")
	// ErrNoStatusSubResource
	ErrNoStatusSubResource = errors.DeclareFormalType(GROUP, ERR_NO_STATUS_SUBRESOURCE, "resource %q has no status sub resource")
)

func New(kind string, msgfmt string, args ...interface{}) error {
	return errors.Newf(GROUP, kind, args, msgfmt, args...)
}

func NewForObject(o interface{}, kind string, msgfmt string, args ...interface{}) error {
	return errors.Newf(GROUP, kind, []interface{}{o}, msgfmt, args...)
}

func NewForObjects(o []interface{}, kind string, msgfmt string, args ...interface{}) error {
	return errors.Newf(GROUP, kind, o, msgfmt, args...)
}

func Wrap(err error, kind string, msgfmt string, args ...interface{}) error {
	return errors.Wrapf(err, GROUP, kind, args, msgfmt, args...)
}

func WrapForObject(err error, o interface{}, kind string, msgfmt string, args ...interface{}) error {
	return errors.Wrapf(err, GROUP, kind, []interface{}{o}, msgfmt, args...)
}

func WrapForObjects(err error, o []interface{}, kind string, msgfmt string, args ...interface{}) error {
	return errors.Wrapf(err, GROUP, kind, o, msgfmt, args...)
}

func NewInvalid(msgfmt string, elem interface{}) error {
	return New(ERR_INVALID, msgfmt, elem)
}
