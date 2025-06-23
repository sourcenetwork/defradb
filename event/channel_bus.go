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
	"sync"
	"sync/atomic"
)

type subscribeCommand *channelSub

type unsubscribeCommand *channelSub

type publishCommand Message

type closeCommand struct{}

// channelSub uses a buffered channel to receive messages.
type channelSub struct {
	id     uint64
	value  chan Message
	events []Name
}

// Message returns the message channel for the subscription.
func (s *channelSub) Message() <-chan Message {
	return s.value
}

// channelBus uses a buffered channel to manage subscribers and publish messages.
type channelBus struct {
	// subID is incremented for each subscriber and used to set subscriber ids.
	subID atomic.Uint64
	// subs is a mapping of subscriber ids to subscriptions.
	subs map[uint64]*channelSub
	// events is a mapping of event names to subscriber ids.
	events map[Name]map[uint64]struct{}
	// commandChannel manages all commands sent to the bufferedBus.
	//
	// It is important that all stuff gets sent through this single channel to ensure
	// that the order of operations is preserved.
	//
	// This does mean that non-event commands can block the database if the buffer
	// size is breached (e.g. if many subscribe commands occupy the buffer).
	commandChannel  chan any
	eventBufferSize int
	hasClosedChan   chan struct{}
	isClosed        bool
	// closeMutex is only locked when the bus is closing.
	closeMutex sync.RWMutex
}

// NewChannelBus creates a new event bus with the given commandBufferSize and
// eventBufferSize.
//
// Should the buffers be filled, subsequent calls on this bus will block.
func NewChannelBus(commandBufferSize int, eventBufferSize int) Bus {
	bus := channelBus{
		subs:            make(map[uint64]*channelSub),
		events:          make(map[Name]map[uint64]struct{}),
		commandChannel:  make(chan any, commandBufferSize),
		hasClosedChan:   make(chan struct{}),
		eventBufferSize: eventBufferSize,
	}
	go bus.handleChannel()
	return &bus
}

// Publish broadcasts the given message to the bus subscribers. Non-blocking.
func (b *channelBus) Publish(msg Message) {
	b.closeMutex.RLock()
	defer b.closeMutex.RUnlock()

	if b.isClosed {
		return
	}
	b.commandChannel <- publishCommand(msg)
}

// Subscribe returns a new subscription that will receive all of the events
// contained in the given list of events.
func (b *channelBus) Subscribe(events ...Name) (Subscription, error) {
	b.closeMutex.RLock()
	defer b.closeMutex.RUnlock()

	if b.isClosed {
		return nil, ErrSubscribedToClosedChan
	}
	sub := &channelSub{
		id:     b.subID.Add(1),
		value:  make(chan Message, b.eventBufferSize),
		events: events,
	}
	b.commandChannel <- subscribeCommand(sub)
	return sub, nil
}

// Unsubscribe removes all event subscriptions and closes the subscription channel.
//
// Will do nothing if this object is already closed.
func (b *channelBus) Unsubscribe(sub Subscription) {
	b.closeMutex.RLock()
	defer b.closeMutex.RUnlock()

	if b.isClosed {
		return
	}
	s, ok := sub.(*channelSub)
	if !ok {
		return
	}
	b.commandChannel <- unsubscribeCommand(s)
}

// Close unsubscribes all active subscribers and closes the command channel.
func (b *channelBus) Close() {
	b.closeMutex.Lock()
	defer b.closeMutex.Unlock()

	if b.isClosed {
		return
	}
	b.isClosed = true
	b.commandChannel <- closeCommand{}
	// Wait for the close command to be handled, in order, before returning
	<-b.hasClosedChan
}

func (b *channelBus) handleChannel() {
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
