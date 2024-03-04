/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package cluster

import (
	restclient "k8s.io/client-go/rest"
)

type APIServerOverride struct{}

var _ Extension = &APIServerOverride{}
var _ RestConfigExtension = &APIServerOverride{}

func (this *APIServerOverride) ExtendConfig(_ Definition, cfg *Config) {
	cfg.AddStringOption(nil, "apiserver-override", "", "", "replace api server url from kubeconfig")
}

func (this *APIServerOverride) Extend(_ Interface, _ *Config) error {
	return nil
}

func (this *APIServerOverride) TweakRestConfig(_ Definition, cfg *Config, restcfg *restclient.Config) error {
	opt := cfg.GetOption("apiserver-override")
	if opt != nil {
		if opt.StringValue() != "" {
			restcfg.Host = opt.StringValue()
		}
	}
	return nil
}
