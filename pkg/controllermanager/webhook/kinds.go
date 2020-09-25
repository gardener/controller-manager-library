/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package webhook

var kindreg = map[WebhookKind]WebhookKindHandlerProvider{}

func RegisterKindHandlerProvider(kind WebhookKind, p WebhookKindHandlerProvider) {
	lock.Lock()
	defer lock.Unlock()
	kindreg[kind] = p
}

func createKindHandlers(ext Environment) (map[WebhookKind]WebhookKindHandler, error) {
	lock.Lock()
	defer lock.Unlock()

	handlers := map[WebhookKind]WebhookKindHandler{}
	for kind, p := range kindreg {
		h, err := p(ext, kind)
		if err != nil {
			return nil, err
		}
		handlers[kind] = h
	}
	return handlers, nil
}
