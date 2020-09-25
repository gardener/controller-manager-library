/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 *
 */

// This package implements a InfoData type for storing certificate sdata in a
// InfoDataList.
// It provided an access interface plus functions to create such a
// Certificate object
//

package certdata

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"

	"github.com/gardener/controller-manager-library/pkg/types/infodata"
)

// Certificate ins the InfoData type to storing Certificates in an InfoDataList
// It implements the InfoData interface plus type specific access functions
type Certificate interface {
	infodata.InfoData

	KeyPair() (tls.Certificate, error)

	Certificates() []*x509.Certificate
	PrivateKey() *rsa.PrivateKey

	CertData() []byte
	KeyData() []byte
}
