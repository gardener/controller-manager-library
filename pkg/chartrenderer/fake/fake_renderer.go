/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package fake

import (
	"github.com/gardener/controller-manager-library/pkg/chartrenderer"
)

// ChartRenderer is a fake renderer for testing
type ChartRenderer struct {
	renderFunc func() (*chartrenderer.RenderedChart, error)
}

// New creates a new Fake chartRenderer
func New(renderFunc func() (*chartrenderer.RenderedChart, error)) chartrenderer.ChartRenderer {
	return &ChartRenderer{
		renderFunc: renderFunc,
	}
}

// Render renders provided chart in struct
func (r *ChartRenderer) Render(_, _, _ string, _ map[string]interface{}) (*chartrenderer.RenderedChart, error) {
	return r.renderFunc()
}
