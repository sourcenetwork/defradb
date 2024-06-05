// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package events

import (
	"sync"
)

// Bus is an event bus used to publish and subscribe to internal
// subsystem messages.
type Bus struct {
	subId  int
	subs   map[int]*Subscription
	events map[string]map[int]any
	mutex  sync.RWMutex
}

// NewBus returns a new event bus.
func NewBus() *Bus {
	return &Bus{
		subs:   make(map[int]*Subscription),
		events: make(map[string]map[int]any),
	}
}

// Publish publishes the given event to all subscribers.
//
// This method will never block if the subscribers buffer is full.
func (b *Bus) Publish(event string, value any) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	subscribers := make(map[int]any)
	// publish to event subscribers
	for id := range b.events[event] {
		subscribers[id] = struct{}{}
	}
	// also publish to wildcard recipients
	for id := range b.events[WildCardEventName] {
		subscribers[id] = struct{}{}
	}
	for id := range subscribers {
		select {
		case b.subs[id].value <- value:
			// published event
		default:
			// channel full
		}
	}
}

// Subscribe returns a new channel that will receive all of the subscribed events.
//
// The size of the buffer should be appropriate for the consumer or events will be dropped.
func (b *Bus) Subscribe(size int, events ...string) *Subscription {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	sub := &Subscription{
		id:     b.subId,
		value:  make(chan any, size),
		events: events,
	}

	b.subId++
	b.subs[sub.id] = sub

	// add sub to all events
	for _, event := range events {
		if _, ok := b.events[event]; !ok {
			b.events[event] = make(map[int]any)
		}
		b.events[event][sub.id] = struct{}{}
	}
	return sub
}

// Unsubscribe unsubscribes from all events and closes the event channel of the given subscription.
func (b *Bus) Unsubscribe(sub *Subscription) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// delete sub from all events
	for _, event := range sub.events {
		delete(b.events[event], sub.id)
	}
	// only close channel once
	if _, ok := b.subs[sub.id]; ok {
		close(sub.value)
	}
	delete(b.subs, sub.id)
}

// Close closes the event bus by unsubscribing all subscribers.
func (b *Bus) Close() {
	var subs []*Subscription

	// get list of all subs
	b.mutex.RLock()
	for _, sub := range b.subs {
		subs = append(subs, sub)
	}
	b.mutex.RUnlock()

	// unsubscribe all subs
	for _, sub := range subs {
		b.Unsubscribe(sub)
	}
}
