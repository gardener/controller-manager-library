/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package conversion

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const FAILURE = "Failure"
const SUCCESS = "Success"

// ErrorResponse creates a new Response for error-handling a request.
func ErrorResponse(req *Request, code int32, err error) *Response {
	var id types.UID
	if req != nil {
		id = req.UID
	}
	return &Response{
		UID: id,
		Result: meta.Status{
			Status:  FAILURE,
			Message: err.Error(),
			Code:    code,
		},
	}
}
