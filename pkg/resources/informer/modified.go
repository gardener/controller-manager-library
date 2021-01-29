/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package informer

import (
	"fmt"

	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"

	"github.com/gardener/controller-manager-library/pkg/utils"
)

type EventHandlerRemovable interface {
	RemoveEventHandler(handler cache.ResourceEventHandler) error
}

type ExtendedSharedIndexInformer interface {
	cache.SharedIndexInformer
	EventHandlerRemovable
	HasEventHandlers() bool
	IsStopped() bool
	IsStarted() bool
}

func (s *sharedIndexInformer) IsStarted() bool {
	s.startedLock.Lock()
	defer s.startedLock.Unlock()
	return s.started
}

func (s *sharedIndexInformer) IsStopped() bool {
	s.startedLock.Lock()
	defer s.startedLock.Unlock()
	return s.stopped
}

func (s *sharedIndexInformer) HasEventHandlers() bool {
	s.startedLock.Lock()
	defer s.startedLock.Unlock()
	if s.stopped {
		return false
	}
	return len(s.processor.listeners) > 0
}

func (s *sharedIndexInformer) RemoveEventHandler(handler cache.ResourceEventHandler) error {
	s.startedLock.Lock()
	defer s.startedLock.Unlock()

	if s.stopped {
		klog.V(2).Infof("Handler %v was not added to shared informer because it has stopped already", handler)
		return nil
	}

	if !utils.IsComparable(handler) {
		return fmt.Errorf("uncomparable handler")
	}

	// in order to safely remove, we have to
	// 1. stop sending add/update/delete notifications
	// 2. remove and stop listener
	// 3. unblock
	s.blockDeltas.Lock()
	defer s.blockDeltas.Unlock()
	s.processor.removeListenerFor(handler)
	return nil
}

func (p *sharedProcessor) removeListenerFor(handler cache.ResourceEventHandler) {
	p.listenersLock.Lock()
	defer p.listenersLock.Unlock()

	listener := p.removeListenerLockedFor(handler)
	if p.listenersStarted && listener != nil {
		close(listener.addCh)
	}
}

func (p *sharedProcessor) removeListenerLockedFor(handler cache.ResourceEventHandler) *processorListener {
	var listener *processorListener
	for i, l := range p.listeners {
		if utils.IsComparable(l.handler) && l.handler == handler {
			listener = l
			p.listeners = append(p.listeners[:i], p.listeners[i+1:]...)
		}
	}
	for i, l := range p.syncingListeners {
		if utils.IsComparable(l.handler) && l.handler == handler {
			p.listeners = append(p.syncingListeners[:i], p.syncingListeners[i+1:]...)
		}
	}
	return listener
}
