/*
 * SPDX-FileCopyrightText: Copyright Contributors to the Gardener project
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

func (this *AbstractObject) GetLabel(name string) string {
	labels := this.GetLabels()
	return labels[name]
}
