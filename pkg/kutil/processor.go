/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package kutil

import (
	"context"
	"time"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/buffer"

	"github.com/gardener/controller-manager-library/pkg/ctxutil"
)

// Modified from kubermetes project (processorListener)

// Processor relays function calls on notifications
// --- using two goroutines, two unbuffered
// channels, and an unbounded ring buffer.  The `add(func)`
// function sends the given notification to `addCh`.  One goroutine
// runs `pop()`, which pumps notifications from `addCh` to `nextCh`
// using storage in the ring buffer while `nextCh` is not keeping up.
// Another goroutine runs `run()`, which receives notifications from
// `nextCh` and synchronously invokes the appropriate handler method.
//
// Processor also keeps track of the adjusted requested resync
// period of the listener.
type Processor struct {
	hndlr  func(interface{})
	group  wait.Group
	ctx    context.Context
	nextCh chan interface{}
	addCh  chan interface{}

	// pendingNotifications is an unbounded ring buffer that holds all notifications not yet distributed.
	// There is one per listener, but a failing/stalled listener will have infinite pendingNotifications
	// added until we OOM.
	// TODO: This is no worse than before, since reflectors were backed by unbounded DeltaFIFOs, but
	// we should try to do something better.
	pendingNotifications buffer.RingGrowing
}

func NewProcessor(ctx context.Context, hndlr func(interface{}), bufferSize int) *Processor {
	p := &Processor{
		ctx:                  ctxutil.CancelContext(ctx),
		hndlr:                hndlr,
		nextCh:               make(chan interface{}),
		addCh:                make(chan interface{}),
		pendingNotifications: *buffer.NewRingGrowing(bufferSize),
	}

	p.group.Start(p.run)
	p.group.Start(p.pop)

	return p
}

func (p *Processor) Add(notification interface{}) {
	p.addCh <- notification
}

func (p *Processor) Stop() {
	ctxutil.Cancel(p.ctx)
}

func (p *Processor) Wait() {
	p.group.Wait()
}

func (p *Processor) pop() {
	defer utilruntime.HandleCrash()
	defer close(p.nextCh) // Tell .run() to stop

	var nextCh chan<- interface{}
	var notification interface{}
	for {
		select {
		case nextCh <- notification:
			// Notification dispatched
			var ok bool
			notification, ok = p.pendingNotifications.ReadOne()
			if !ok { // Nothing to pop
				nextCh = nil // Disable this select case
			}
		case notificationToAdd, ok := <-p.addCh:
			if !ok {
				return
			}
			if notification == nil { // No notification to pop (and pendingNotifications is empty)
				// Optimize the case - skip adding to pendingNotifications
				notification = notificationToAdd
				nextCh = p.nextCh
			} else { // There is already a notification waiting to be dispatched
				p.pendingNotifications.WriteOne(notificationToAdd)
			}
		case <-p.ctx.Done():
			close(p.addCh)
			return
		}
	}
}

func (p *Processor) run() {
	// this call blocks until the channel is closed.  When a panic happens during the notification
	// we will catch it, **the offending item will be skipped!**, and after a short delay (one second)
	// the next notification will be attempted.  This is usually better than the alternative of never
	// delivering again.
	stopCh := make(chan struct{})
	wait.Until(func() {
		for next := range p.nextCh {
			p.hndlr(next)
		}
		// the only way to get here is if the p.nextCh is empty and closed
		close(stopCh)
	}, 1*time.Second, stopCh)
}
