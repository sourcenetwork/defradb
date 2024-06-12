// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package event

import (
	"sync/atomic"
)

type subscribeCommand *Subscription

type unsubscribeCommand *Subscription

type publishCommand Message

type closeCommand struct{}

// bufferedBus must implement the Bus interface
var _ (Bus) = (*bufferedBus)(nil)

// bufferedBus is a bus that uses a buffered channel to manage subscribers and publish messages.
type bufferedBus struct {
	// subID is incremented for each subscriber and used to set subscriber ids.
	subID atomic.Uint64
	// subs is a mapping of subscriber ids to subscriptions.
	subs map[uint64]*Subscription
	// events is a mapping of event names to subscriber ids.
	events map[Name]map[uint64]struct{}
	// commandChannel manages all commands sent to this simpleChannel.
	//
	// It is important that all stuff gets sent through this single channel to ensure
	// that the order of operations is preserved.
	//
	// WARNING: This does mean that non-event commands can block the database if the buffer
	// size is breached (e.g. if many subscribe commands occupy the buffer).
	commandChannel  chan any
	eventBufferSize int
	hasClosedChan   chan struct{}
	isClosed        atomic.Bool
}

// NewBufferedBus creates a new event bus with the given commandBufferSize and
// eventBufferSize.
//
// Should the buffers be filled, subsequent calls on this bus will block.
func NewBufferedBus(commandBufferSize int, eventBufferSize int) *bufferedBus {
	bus := bufferedBus{
		subs:            make(map[uint64]*Subscription),
		events:          make(map[Name]map[uint64]struct{}),
		commandChannel:  make(chan any, commandBufferSize),
		hasClosedChan:   make(chan struct{}),
		eventBufferSize: eventBufferSize,
	}
	go bus.handleChannel()
	return &bus
}

func (b *bufferedBus) Publish(msg Message) {
	if b.isClosed.Load() {
		return
	}
	b.commandChannel <- publishCommand(msg)
}

func (b *bufferedBus) Subscribe(events ...Name) (*Subscription, error) {
	if b.isClosed.Load() {
		return nil, ErrSubscribedToClosedChan
	}
	sub := &Subscription{
		id:     b.subID.Add(1),
		value:  make(chan Message, b.eventBufferSize),
		events: events,
	}
	b.commandChannel <- subscribeCommand(sub)
	return sub, nil
}

func (b *bufferedBus) Unsubscribe(sub *Subscription) {
	if b.isClosed.Load() {
		return
	}
	b.commandChannel <- unsubscribeCommand(sub)
}

func (b *bufferedBus) Close() {
	if b.isClosed.Load() {
		return
	}
	b.isClosed.Store(true)
	b.commandChannel <- closeCommand{}
	// Wait for the close command to be handled, in order, before returning
	<-b.hasClosedChan
}

func (b *bufferedBus) handleChannel() {
	for cmd := range b.commandChannel {
		switch t := cmd.(type) {
		case closeCommand:
			for _, subscriber := range b.subs {
				close(subscriber.value)
			}
			close(b.commandChannel)
			close(b.hasClosedChan)
			return

		case subscribeCommand:
			for _, event := range t.events {
				if _, ok := b.events[event]; !ok {
					b.events[event] = make(map[uint64]struct{})
				}
				b.events[event][t.id] = struct{}{}
			}
			b.subs[t.id] = t

		case unsubscribeCommand:
			if _, ok := b.subs[t.id]; !ok {
				continue // not subscribed
			}
			for _, event := range t.events {
				delete(b.events[event], t.id)
			}
			delete(b.subs, t.id)
			close(t.value)

		case publishCommand:
			for id := range b.events[WildCardName] {
				b.subs[id].value <- Message(t)
			}
			for id := range b.events[t.Name] {
				if _, ok := b.events[WildCardName][id]; ok {
					continue
				}
				b.subs[id].value <- Message(t)
			}
		}
	}
}
