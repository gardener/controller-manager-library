/*
 * SPDX-FileCopyrightText: Contributors to the Gardener project
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package abstract

func hasFinalizer(key string, obj ObjectData) bool {
	for _, name := range obj.GetFinalizers() {
		if name == key {
			return true
		}
	}
	return false
}

func (this *AbstractObject) HasFinalizer(key string) bool {
	for _, name := range this.GetFinalizers() {
		if name == key {
			return true
		}
	}
	return false
}

func (this *AbstractObject) SetFinalizer(key string) error {
	if !hasFinalizer(key, this.ObjectData) {
		this.SetFinalizers(append(this.GetFinalizers(), key))
	}
	return nil
}

func (this *AbstractObject) RemoveFinalizer(key string) error {
	list := this.GetFinalizers()
	for i, name := range list {
		if name == key {
			this.SetFinalizers(append(list[:i], list[i+1:]...))
			return nil
		}
	}
	return nil
}
