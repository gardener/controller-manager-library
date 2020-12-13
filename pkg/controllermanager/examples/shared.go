/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package examples

import (
	"github.com/gardener/controller-manager-library/pkg/controllermanager/extension"
	"github.com/gardener/controller-manager-library/pkg/ctxutil"
	"github.com/gardener/controller-manager-library/pkg/types/infodata/simple"
)

var key = ctxutil.SimpleKey("shared")

type SharedInfo struct {
	Values simple.Values
}

func GetOrCreateShared(env extension.Environment) *SharedInfo {
	return env.ControllerManager().GetOrCreateSharedValue(key, func() interface{} {
		return &SharedInfo{Values: simple.Values{}}
	}).(*SharedInfo)
}

func GetShared(env extension.Environment) *SharedInfo {
	s := env.ControllerManager().GetSharedValue(key)
	if s == nil {
		return nil
	}
	return s.(*SharedInfo)
}
