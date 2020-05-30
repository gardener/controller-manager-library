/*
 * Copyright 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */

package cert

import (
	"context"
	"fmt"

	certsecret "github.com/gardener/controller-manager-library/pkg/certmgmt/secret"
	"github.com/gardener/controller-manager-library/pkg/certs"
	"github.com/gardener/controller-manager-library/pkg/config"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/extension"
	"github.com/gardener/controller-manager-library/pkg/logger"
)

type CertConfig struct {
	name   string
	prefix string

	CommonName   string
	Organization string

	Secret     string
	Service    string
	Hostnames  []string
	CertFile   string
	KeyFile    string
	CACertFile string
	CAKeyFile  string
}

func (this *CertConfig) AddOptionsToSet(set config.OptionSet) {
	set.AddStringOption(&this.Secret, this.prefix+"secret", "", "", fmt.Sprintf("name of secret to maintain for %s server", this.name))
	set.AddStringOption(&this.Service, this.prefix+"service", "", "", fmt.Sprintf("name of service to use for %s server", this.name))
	set.AddStringArrayOption(&this.Hostnames, this.prefix+"hostname", "", nil, fmt.Sprintf("hostname to use for %s registration", this.name))
	set.AddStringOption(&this.CertFile, this.prefix+"certfile", "", "", fmt.Sprintf("%s server certificate file", this.name))
	set.AddStringOption(&this.KeyFile, this.prefix+"keyfile", "", "", fmt.Sprintf("%s server certificate key file", this.name))
	set.AddStringOption(&this.CACertFile, this.prefix+"cacertfile", "", "", fmt.Sprintf("%s server ca certificate file", this.name))
	set.AddStringOption(&this.CAKeyFile, this.prefix+"cakeyfile", "", "", fmt.Sprintf("%s server ca certificate key file", this.name))
}

func OptionSourceCreator(name, prefix string, common, org string) extension.OptionSourceCreator {
	return func() config.OptionSource {
		cfg := NewCertConfig(name, prefix)
		cfg.CommonName = common
		cfg.Organization = org
		return cfg
	}
}

func NewCertConfig(name, prefix string) *CertConfig {
	return &CertConfig{name: name, prefix: prefix}
}

func (this *CertConfig) Used() bool {
	if len(this.Hostnames) > 0 {
		return true
	}
	if this.Service != "" {
		return true
	}
	return false
}

func (this *CertConfig) CreateAccess(ctx context.Context, logger logger.LogContext, cluster cluster.Interface, namespace string, keys ...certsecret.Keys) (certs.CertificateSource, error) {
	if this.CertFile != "" {
		return CreateFileCertificateSource(ctx, logger, this)
	}
	if this.Secret != "" {
		return CreateSecretCertificateSource(ctx, logger, cluster, namespace, this, keys...)
	}
	return nil, nil
}
