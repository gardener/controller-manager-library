/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package handler

/*
  Creation Flow for Modules

  a) Module Creation Time: Handler Creation by calling all HandlerType function
  b) Calling Setup on all Handler
  c) Calling Start on all Handler
*/

type Interface interface {
}

type SetupInterface interface {
	Setup() error
}

type StartInterface interface {
	Start() error
}

type LegacyInterface interface {
	Interface
	SetupInterface
	StartInterface
}
