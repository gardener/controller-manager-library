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

import (
	"fmt"
	"github.com/pkg/errors"
	"io"
)

/*
 * This package offers categorized error objects, that provide
 * group and kind information, to formally handle different kinds
 * of errors. Additionally such an error may offer a defined
 * set of values specific for the error kind, that is available for
 * the error handler. The meaning and type of those objects and their order
 * should always be the same, regardles of the way the error object is
 * generated. It might be ok to omit the values at all.
 *
 * All errors provide a stacktrace using the StackTracer interface.
 * It provides information about the error location in the code and the
 * call history. Using the Wrap variant for creating errors it is
 * possible to keep the error history when creating errors for errors
 * provided by nested function calls.
 *
 * Using the functions Newf/Wrapf it is possible to create such an error
 * on the fly by specifying all the relevant values. A better way is
 * to use predefined error types. They support the creation of a dedicated
 * kind of error. There two kinds of error types
 *
 * - Type created by DeclareType just contains the category meta data for
 *         Group and Kind
 * - Formal created by DeclareFormalType addtionally defines the standard error
 *         message format based on arguments.
 *         Those arguments will automatically be used as provided
 *         error objects,
 *
 * The error objects offer different formatting options for the fmt package
 * - %s  just the categaory + error message
 * - %q  quoted error message
 * - %v  error history
 * - %-v additional error location
 * - %+v additional error call stack
 */

type Categorized interface {
	error
	Group() string
	Kind() string
	Cause() error
}

type Formal interface {
	Categorized
	Arg(n int) interface{}
	Length() int
}

type GroupKind struct {
	group string
	kind  string
}

func (this *GroupKind) Group() string {
	return this.group
}

func (this *GroupKind) Kind() string {
	return this.kind
}

type Type struct {
	GroupKind
}

type FormalType struct {
	Type
	format string
}

func DeclareType(group, kind string) *Type {
	return &Type{
		GroupKind: GroupKind{
			group: group,
			kind:  kind,
		},
	}
}

func DeclareFormalType(group, kind, format string) *FormalType {
	return &FormalType{
		Type: Type{
			GroupKind: GroupKind{
				group: group,
				kind:  kind,
			},
		},
		format: format,
	}
}

type fundamental struct {
	error
	etype *Type
	args  []interface{}
}

type StackTracer interface {
	StackTrace() errors.StackTrace
}

func (this *fundamental) Group() string {
	return this.etype.group
}
func (this *fundamental) Kind() string {
	return this.etype.kind
}

func (this *fundamental) Error() string {
	return fmt.Sprintf("%s/%s: %s", this.Group(), this.Kind(), this.error.Error())
}

func (this *fundamental) StackTrace() errors.StackTrace {
	return this.error.(StackTracer).StackTrace()[1:]
}

func (this *fundamental) Arg(n int) interface{} {
	if len(this.args) >= n-1 {
		return this.args[n]
	}
	return nil
}

func (this *fundamental) Length() int {
	return len(this.args)
}
func (this *fundamental) Cause() error {
	return nil
}

func (this *fundamental) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			io.WriteString(s, this.Error())
			fmt.Fprintf(s, "%+v", this.StackTrace())
			return
		}
		if s.Flag('-') {
			io.WriteString(s, this.Error())
			fmt.Fprintf(s, "\n%+v", this.StackTrace()[0])
			return
		}
		fallthrough
	case 's':
		io.WriteString(s, this.Error())
	case 'q':
		fmt.Fprintf(s, "%q", this.Error())
	}
}

func (this *FormalType) New(args ...interface{}) Categorized {
	err := errors.Errorf(this.format, args...)
	return &fundamental{
		etype: &this.Type,
		args:  args,
		error: err,
	}
}

func (this *FormalType) Newf(objs []interface{}, format string, args ...interface{}) Categorized {
	err := errors.Errorf(format, args...)
	return &fundamental{
		etype: &this.Type,
		args:  objs,
		error: err,
	}
}

func Newf(group, kind string, objs []interface{}, format string, args ...interface{}) Categorized {
	err := errors.Errorf(format, args...)
	return &fundamental{
		etype: &Type{
			GroupKind{
				group: group,
				kind:  kind,
			},
		},
		args:  objs,
		error: err,
	}
}

func Wrapf(cause error, group, kind string, objs []interface{}, format string, args ...interface{}) Categorized {
	err := errors.Errorf(format, args...)
	return &withCause{
		fundamental: fundamental{
			etype: &Type{
				GroupKind{
					group: group,
					kind:  kind,
				},
			},
			args:  objs,
			error: err,
		},
		cause: cause,
	}
}

func (this *Type) Newf(format string, args ...interface{}) Categorized {
	err := errors.Errorf(format, args...)
	return &fundamental{
		etype: this,
		args:  nil,
		error: err,
	}
}

type withCause struct {
	fundamental
	cause error
}

func (this *FormalType) Wrap(cause error, args ...interface{}) Categorized {
	err := errors.Errorf(this.format, args...)
	return &withCause{
		fundamental: fundamental{
			etype: &this.Type,
			args:  args,
			error: err,
		},
		cause: cause,
	}
}

func (this *Type) Wrapf(cause error, format string, args ...interface{}) Categorized {
	err := errors.Errorf(format, args...)
	return &withCause{
		fundamental: fundamental{
			etype: this,
			args:  nil,
			error: err,
		},
		cause: cause,
	}
}

func (this *withCause) Cause() error {
	return this.cause
}

func (this *withCause) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		this.fundamental.Format(s, verb)
		if this.cause != nil {
			io.WriteString(s, "\ncaused by: ")
			if c, ok := this.cause.(fmt.Formatter); ok {
				c.Format(s, verb)
			} else {
				io.WriteString(s, this.cause.Error())
			}
		}
	case 's':
		io.WriteString(s, fmt.Sprintf("%s/%s: ", this.Group(), this.Kind()))
		io.WriteString(s, this.error.Error())
	case 'q':
		fmt.Fprintf(s, "%q", this.error.Error())
	}
}
