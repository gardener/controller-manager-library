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

package cert

import (
	"context"
	"fmt"
	"time"

	"github.com/gardener/controller-manager-library/pkg/certmgmt"
	certsecret "github.com/gardener/controller-manager-library/pkg/certmgmt/secret"
	"github.com/gardener/controller-manager-library/pkg/certs"
	"github.com/gardener/controller-manager-library/pkg/certs/access"
	"github.com/gardener/controller-manager-library/pkg/certs/file"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
)

func CreateSecretCertificateSource(ctx context.Context, logger logger.LogContext, cluster cluster.Interface, namespace string, cfg *CertConfig, keys ...certsecret.Keys) (certs.CertificateSource, error) {
	secret := certsecret.NewSecret(cluster, resources.NewObjectName(namespace, cfg.Secret), keys...)
	hosts := certmgmt.NewCompoundHosts()
	for _, h := range cfg.Hostnames {
		logger.Infof("using hostname for certificate: %s", h)
		hosts.Add(certmgmt.NewDNSName(h))
	}
	if cfg.Service != "" {
		logger.Infof("using service for certificate: %s/%s", cfg.Service, namespace)
		hosts.Add(certmgmt.NewServiceHosts(cfg.Service, namespace))
	}

	logger.Infof("using certificate for ips: %v, dns: %v", hosts.GetIPs(), hosts.GetDNSNames())

	var certcfg *certmgmt.Config
	if cfg.CommonName != "" {
		if len(hosts) == 0 {
			return nil, fmt.Errorf("hosts for managed certificate secret required")
		}
		logger.Infof("managing certificate")
		certcfg = &certmgmt.Config{
			CommonName:   cfg.CommonName,
			Organization: []string{cfg.Organization},
			Validity:     10 * 24 * time.Hour,
			Rest:         24 * time.Hour,
			Hosts:        hosts,
		}
	} else {
		logger.Infof("externally managed certificate")
	}
	return access.New(ctx, logger, secret, certcfg)
}

func CreateFileCertificateSource(ctx context.Context, logger logger.LogContext, cfg *CertConfig) (certs.CertificateSource, error) {
	return file.New(ctx, logger, cfg.CertFile, cfg.KeyFile, cfg.CACertFile, cfg.CAKeyFile)
}
