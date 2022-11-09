// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package events

import "sync"

type Publisher[T any] struct {
	ch chan T

	closingCh chan struct{}
	isClosed  bool
	writersWG sync.WaitGroup
	syncLock  sync.Mutex
}

// NewPublisher creates a Publisher
func NewPublisher[T any](ch chan T) *Publisher[T] {
	return &Publisher[T]{
		ch:        ch,
		closingCh: make(chan struct{}),
	}
}

// Read returns the channel to write
func (p *Publisher[T]) Read() <-chan T {
	return p.ch
}

// Write into the channel in a different goroutine
func (p *Publisher[T]) Write(data T) {
	go func(data T) {
		p.syncLock.Lock()
		p.writersWG.Add(1)
		p.syncLock.Unlock()
		defer p.writersWG.Done()

		select {
		case <-p.closingCh:
			return
		default:
		}

		select {
		case <-p.closingCh:
		case p.ch <- data:
		}
	}(data)
}

// Closes channel, draining any blocked writes
func (p *Publisher[T]) Close() {
	close(p.closingCh)

	go func() {
		for range p.ch {
		}
	}()

	p.syncLock.Lock()
	p.writersWG.Wait()
	p.isClosed = true
	close(p.ch)
	p.syncLock.Unlock()
}

// Closes channel, draining any blocked writes
func (p *Publisher[T]) IsClosed() bool {
	p.syncLock.Lock()
	defer p.syncLock.Unlock()

	return p.isClosed
}
