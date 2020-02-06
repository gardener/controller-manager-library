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

	"github.com/gardener/controller-manager-library/pkg/infodata"
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
