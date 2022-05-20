/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package extension

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConfigSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Extension Suite")
}
