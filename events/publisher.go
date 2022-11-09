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

// Publisher holds a channel and sync mechanism that enable safe writes
// where the reader is expected to be the one closing the channel.
type Publisher struct {
	ch chan any

	closingCh chan struct{}
	isClosed  bool
	writersWG sync.WaitGroup
	syncLock  sync.Mutex
}

// NewPublisher creates a Publisher with the given channel
func NewPublisher(ch chan any) *Publisher {
	return &Publisher{
		ch:        ch,
		closingCh: make(chan struct{}),
	}
}

// Read returns the channel
func (p *Publisher) Read() <-chan any {
	return p.ch
}

// Write into the channel
func (p *Publisher) Write(data any) {
	go func(data any) {
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
func (p *Publisher) Close() {
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

// IsClosed returns true if the channel has been closed.
func (p *Publisher) IsClosed() bool {
	p.syncLock.Lock()
	defer p.syncLock.Unlock()

	return p.isClosed
}
