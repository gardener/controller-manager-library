/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package v1alpha1

import (
	"fmt"
	"net/url"
	"strconv"

	"k8s.io/apimachinery/pkg/conversion"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/examples/apis/example"
)

func Convert_v1alpha1_ExampleSpec_To_example_ExampleSpec(in *ExampleSpec, out *example.ExampleSpec, s conversion.Scope) error {
	if in.Port > 0 {
		out.URL = fmt.Sprintf("%s://%s:%d/%s", in.URLScheme, in.Hostname, in.Port, in.Path)
	} else {
		out.URL = fmt.Sprintf("%s://%s/%s", in.URLScheme, in.Hostname, in.Path)
	}
	return nil
}

func Convert_example_ExampleSpec_To_v1alpha1_ExampleSpec(in *example.ExampleSpec, out *ExampleSpec, s conversion.Scope) error {
	u, err := url.Parse(in.URL)
	if err == nil {
		out.URLScheme = u.Scheme
		out.Hostname = u.Hostname()
		out.Path = u.Path
		out.Port, err = strconv.Atoi(u.Port())
		if err != nil {
			out.Port = 0
			err = nil
		}
	}
	return err
}
