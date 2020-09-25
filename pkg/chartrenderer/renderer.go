/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package chartrenderer

// ChartRenderer is an interface for rendering Helm Charts from path, name, namespace and values.
type ChartRenderer interface {
	Render(chartPath, releaseName, namespace string, values map[string]interface{}) (*RenderedChart, error)
}

// RenderedChart holds a map of rendered templates file with template file name as key and
// rendered template as value.
type RenderedChart struct {
	ChartName string
	Files     map[string]string
}
