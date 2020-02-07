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
