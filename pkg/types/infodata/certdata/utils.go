/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 *
 */

package certdata

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"github.com/gardener/controller-manager-library/pkg/utils/pkiutil"
)

func encodePrivateKeyPEM(key *rsa.PrivateKey) []byte {
	block := pem.Block{
		Type:  pkiutil.RSAPrivateKeyBlockType,
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}
	return pem.EncodeToMemory(&block)
}

func encodeCertsPEM(certs []*x509.Certificate) []byte {
	bytes := []byte{}
	for _, cert := range certs {
		bytes = append(bytes, pkiutil.EncodeCertPEM(cert)...)
	}
	return bytes
}
