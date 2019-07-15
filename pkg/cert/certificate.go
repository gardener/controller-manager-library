/*
 * Copyright 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"
)

type info struct {
	cert   []byte
	key    []byte
	cacert []byte
	cakey  []byte
}

func (this *info) Cert() []byte {
	return this.cert
}

func (this *info) CACert() []byte {
	return this.cacert
}

func (this *info) Key() []byte {
	return this.key
}

func (this *info) CAKey() []byte {
	return this.cakey
}

func NewCertInfo(cert []byte, key []byte, cacert []byte, cakey []byte) CertificateInfo {
	return &info{
		cert:   cert,
		key:    key,
		cacert: cacert,
		cakey:  cakey,
	}
}

func newPrivateKey() (*rsa.PrivateKey, error) {
	signer, err := pkiutil.NewPrivateKey()
	if err != nil {
		return nil, err
	}
	key, ok := signer.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not a private key: %t", key)
	}
	return key, nil
}

// EncodePrivateKeyPEM returns PEM-encoded private key data
func encodePrivateKeyPEM(key *rsa.PrivateKey) []byte {
	block := pem.Block{
		Type:  pkiutil.RSAPrivateKeyBlockType,
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}
	return pem.EncodeToMemory(&block)
}

func UpdateCertificate(old CertificateInfo, commonname, dnsname string, duration time.Duration) (CertificateInfo, error) {
	new := &info{}
	if old != nil {
		new.cert = old.Cert()
		new.key = old.Key()
		new.cacert = old.CACert()
		new.cakey = old.CAKey()
	}

	var caKey *rsa.PrivateKey
	var caCert *x509.Certificate
	var newKey *rsa.PrivateKey
	var newCert *x509.Certificate
	var err error
	var ok bool

	if !IsValid(new, dnsname, duration) {
		fmt.Printf("not valid\n")
		if new.cacert != nil {
			fmt.Printf("cacert found\n")
			ok = Valid(new.cakey, new.cacert, new.cacert, "", 5*time.Hour*24)
			if ok {
				fmt.Printf("cacert not valid\n")
				k, err := keyutil.ParsePrivateKeyPEM(new.cakey)
				if err != nil {
					ok = false
				} else {
					caKey, ok = k.(*rsa.PrivateKey)
				}
				certs, err := cert.ParseCertsPEM(new.cacert)
				if err != nil {
					ok = false
				} else {
					caCert = certs[0]
				}
			}
		}
		if new.cacert == nil || !ok {
			fmt.Printf("generate cacert\n")

			caKey, err = newPrivateKey()
			if err != nil {
				return nil, fmt.Errorf("failed to create the CA key pair: %s", err)
			}
			new.cakey = encodePrivateKeyPEM(caKey)
			caCert, err = cert.NewSelfSignedCACert(cert.Config{CommonName: "webhook-cert-ca:" + commonname}, caKey)
			if err != nil {
				return nil, fmt.Errorf("failed to create the CA cert: %s", err)
			}
			new.cacert = pkiutil.EncodeCertPEM(caCert)
		}

		fmt.Printf("generate key\n")
		newKey, err = newPrivateKey()
		if err != nil {
			return nil, fmt.Errorf("failed to create the server key pair: %s", err)
		}
		new.key = encodePrivateKeyPEM(newKey)
		fmt.Printf("generate cert\n")
		newCert, err = pkiutil.NewSignedCert(
			&cert.Config{
				CommonName: "client:" + commonname,
				AltNames: cert.AltNames{
					DNSNames: []string{dnsname},
				},
				Usages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			},
			newKey, caCert, caKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create the server cert: %s", err)
		}
		new.cert = pkiutil.EncodeCertPEM(newCert)
		return new, nil
	}
	return old, nil
}

func IsValid(info CertificateInfo, dnsname string, duration time.Duration) bool {
	if info.Cert() == nil || info.Key() == nil {
		fmt.Printf("cert or key not set\n")
		return false
	}
	if info.CACert() == nil {
		fmt.Printf("cacert not set\n")
		return false
	}
	return Valid(info.Key(), info.Cert(), info.CACert(), dnsname, duration)
}

func Valid(key []byte, cert []byte, cacert []byte, dnsname string, duration time.Duration) bool {

	if len(cert) == 0 || len(key) == 0 || len(cacert) == 0 {
		fmt.Printf("something empty\n")
		return false
	}

	_, err := tls.X509KeyPair(cert, key)
	if err != nil {
		fmt.Printf("key does not match cert\n")
		return false
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(cacert) {
		fmt.Printf("cannot create pool\n")
		return false
	}
	block, _ := pem.Decode([]byte(cert))
	if block == nil {
		fmt.Printf("cannot decode cert\n")
		return false
	}
	c, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		fmt.Printf("cannot parse cert\n")
		return false
	}
	ops := x509.VerifyOptions{
		DNSName:     dnsname,
		Roots:       pool,
		CurrentTime: time.Now().Add(duration),
	}
	_, err = c.Verify(ops)
	fmt.Printf("val: %s\n", err)
	return err == nil
}
