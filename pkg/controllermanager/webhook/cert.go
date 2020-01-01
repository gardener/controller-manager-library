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

package webhook

import (
	"context"
	"fmt"
	"time"

	"github.com/gardener/controller-manager-library/pkg/certmgmt"
	certsecret "github.com/gardener/controller-manager-library/pkg/certmgmt/secret"
	"github.com/gardener/controller-manager-library/pkg/certs"
	"github.com/gardener/controller-manager-library/pkg/certs/access"
	"github.com/gardener/controller-manager-library/pkg/certs/file"
	"github.com/gardener/controller-manager-library/pkg/resources"
)

func CreateSecretCertificateSource(ctx context.Context, ext Environment) (certs.CertificateSource, error) {
	cfg := ext.GetConfig()
	cluster := ext.GetCluster(cfg.Cluster)
	if cluster == nil {
		return nil, fmt.Errorf("cluster %q for webhook server secret not found", cfg.Cluster)
	}
	secret := certsecret.NewSecret(cluster, resources.NewObjectName(ext.Namespace(), cfg.Secret))
	hosts := certmgmt.NewCompoundHosts()
	if cfg.Hostname != "" {
		ext.Infof("using hostname for certificate: %s", cfg.Hostname)
		hosts.Add(certmgmt.NewDNSName(cfg.Hostname))
	}
	if cfg.Service != "" {
		ext.Infof("using service for certificate: %s/%s", cfg.Service, ext.Namespace())
		hosts.Add(certmgmt.NewServiceHosts(cfg.Service, ext.Namespace()))
	}

	ext.Infof("using certificate for ips: %v, dns: %v", hosts.GetIPs(), hosts.GetDNSNames())
	certcfg := &certmgmt.Config{
		CommonName:   ext.Name(),
		Organization: []string{"kubernetes"},
		Validity:     10 * 24 * time.Hour,
		Rest:         24 * time.Hour,
		Hosts:        hosts,
	}
	return access.New(ctx, ext, secret, certcfg)
}

func CreateFileCertificateSource(ctx context.Context, ext Environment) (certs.CertificateSource, error) {
	cfg := ext.GetConfig()
	return file.New(ctx, ext, cfg.CertFile, cfg.KeyFile, cfg.CACertFile, cfg.CAKeyFile)
}
