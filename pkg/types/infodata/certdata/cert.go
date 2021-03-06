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
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"

	"k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"

	"github.com/gardener/controller-manager-library/pkg/types/infodata"
)

// certificate is the actual implementation for a Certificate object
// regardless of the version stored in a InfoDataList
// It always marshals itself to the generic version because dedicated
// versions are not expected. If this is not expected, a dedicated version
// should be used for marshalling instead.
type _certificate struct {
	certdata []byte
	keydata  []byte
	certs    []*x509.Certificate
	key      *rsa.PrivateKey
}

var _ Certificate = &_certificate{}

// NewCertificateByData is a constructor for a Certificate Type
// using pem representations
func NewCertificateByData(cert, key []byte) (Certificate, error) {
	this := &_certificate{certdata: cert, keydata: key}
	return this.setup()
}

// NewCertificate is a constructor for a Certificate Type
// using the semantical content
func NewCertificate(key *rsa.PrivateKey, certs ...*x509.Certificate) (Certificate, error) {
	this := &_certificate{certs: certs, key: key}
	return this.setup()
}

func (this *_certificate) TypeVersion() infodata.TypeVersion {
	return T_CERTIFICATE
}

func (this *_certificate) Marshal() ([]byte, error) {
	return json.Marshal(&certificateV1{Certificate: string(this.certdata), Key: string(this.KeyData())})
}

func (this *_certificate) setup() (*_certificate, error) {
	var err error
	if this.certs == nil {
		this.certs, err = cert.ParseCertsPEM(this.certdata)
		if err != nil {
			return nil, err
		}
	}
	if this.key == nil {
		k, err := keyutil.ParsePrivateKeyPEM(this.keydata)
		if err != nil {
			return nil, err
		}
		key, ok := k.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("no rsa key")
		}
		this.key = key
	}
	if this.certdata == nil {
		this.certdata = encodeCertsPEM(this.certs)
	}
	if this.keydata == nil {
		this.keydata = encodePrivateKeyPEM(this.key)
	}
	return this, nil
}

func (this *_certificate) Certificates() []*x509.Certificate {
	return this.certs
}

func (this *_certificate) PrivateKey() *rsa.PrivateKey {
	return this.key
}

func (this *_certificate) KeyPair() (tls.Certificate, error) {
	return tls.X509KeyPair(this.CertData(), this.KeyData())
}

func (this *_certificate) CertData() []byte {
	return this.certdata
}

func (this *_certificate) KeyData() []byte {
	return this.keydata
}
