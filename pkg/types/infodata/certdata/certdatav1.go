/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 *
 */

package certdata

import (
	"encoding/json"
	"fmt"

	"github.com/gardener/controller-manager-library/pkg/types/infodata"
)

const T_CERTIFICATE = infodata.TypeVersion("Certificate") // "Certificate/v1"

// init registers the supported version v1 for the InfoData type Certificate
// because the representation will basically not change there is a
// default unversioned type name.
// If the structure is potetially volatile there should always be a version
// suffix for the typeversion (like Certificate/v1)
func init() {
	infodata.Register(T_CERTIFICATE, unmarshalV1)
}

type certificateV1 struct {
	Certificate string `json:"certificate"`
	Key         string `json:"privateKey"`
}

// unmarshalV1 creates Certificate object from version v1 for the type
// Certificate
func unmarshalV1(bytes []byte) (infodata.InfoData, error) {
	if bytes == nil {
		return nil, fmt.Errorf("no data given")
	}
	data := &certificateV1{}
	err := json.Unmarshal(bytes, data)
	if err != nil {
		return nil, err
	}
	return NewCertificateByData([]byte(data.Certificate), []byte(data.Key))
}
