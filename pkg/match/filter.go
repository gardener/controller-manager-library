/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package match

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/gardener/controller-manager-library/pkg/fieldpath"
)

type Matcher interface {
	Match(obj interface{}) bool
}

////////////////////////////////////////////////////////////////////////////////

func FilterList(list interface{}, matchers ...Matcher) interface{} {
	var m Matcher
	if len(matchers) == 1 {
		m = matchers[0]
	} else {
		m = And(matchers...)
	}
	value := reflect.ValueOf(list)
	for value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if value.Kind() == reflect.Array || value.Kind() == reflect.Slice {
		n := reflect.New(reflect.SliceOf(value.Type().Elem())).Elem()
		l := value.Len()
		for i := 0; i < l; i++ {
			if m.Match(value.Index(i).Interface()) {
				n = reflect.Append(n, value.Index(i))
			}
		}
		return n.Interface()
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type MatchFunc func(obj interface{}) bool

func (this MatchFunc) Match(obj interface{}) bool {
	return this(obj)
}

var All = MatchFunc(func(interface{}) bool { return true })
var None = MatchFunc(func(interface{}) bool { return false })

////////////////////////////////////////////////////////////////////////////////

func Or(filters ...Matcher) Matcher {
	return MatchFunc(func(obj interface{}) bool {
		for _, f := range filters {
			if f.Match(obj) {
				return true
			}
		}
		return false
	})
}

func And(filters ...Matcher) Matcher {
	return MatchFunc(func(obj interface{}) bool {
		if len(filters) == 0 {
			return false
		}
		for _, f := range filters {
			if !f.Match(obj) {
				return false
			}
		}
		return true
	})
}

func Not(filter Matcher) Matcher {
	return MatchFunc(func(obj interface{}) bool {
		return !filter.Match(obj)
	})
}

////////////////////////////////////////////////////////////////////////////////

func MatchFieldValueByName(path string, value interface{}) Matcher {
	return MatchFieldValue(fieldpath.MustFieldPath(path), value)
}

func MatchFieldValuesByName(path string, values ...interface{}) Matcher {
	return MatchFieldValues(fieldpath.MustFieldPath(path), values...)
}

func MatchFieldValue(f fieldpath.ValueGetter, value interface{}) Matcher {
	return MatchFunc(func(obj interface{}) bool {
		v, err := f.Get(obj)
		return err == nil && reflect.DeepEqual(v, value)
	})
}

func MatchFieldValues(f fieldpath.ValueGetter, values ...interface{}) Matcher {
	return MatchFunc(func(obj interface{}) bool {
		v, err := f.Get(obj)
		if err != nil {
			return false
		}
		for _, value := range values {
			if !reflect.DeepEqual(v, value) {
				return false
			}
		}
		return true
	})
}

func MatchFieldPattern(f fieldpath.ValueGetter, pattern string) Matcher {
	pat, err := regexp.Compile(pattern)
	if err != nil {
		panic(fmt.Sprintf("invalid pattern %q: %s", pattern, err))
	}
	return MatchFunc(func(obj interface{}) bool {
		v, err := f.Get(obj)
		if err != nil {
			return false
		}
		s, ok := v.(string)
		return ok && pat.MatchString(s)
	})
}
