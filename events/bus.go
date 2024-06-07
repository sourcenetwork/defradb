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

// Message contains event info.
type Message struct {
	// Name is the name of the event this message was generated from.
	Name string
	// Data contains optional event information.
	Data any
}

// NewMessage returns a new message with the given name and optional data.
func NewMessage(name string, data any) Message {
	return Message{name, data}
}

// Subscription is a read-only event stream.
type Subscription struct {
	id     int
	value  chan Message
	events []string
}

// Message returns the next event value from the subscription.
func (s *Subscription) Message() <-chan Message {
	return s.value
}

// Events returns the names of all subscribed events.
func (s *Subscription) Events() []string {
	return s.events
}

// Bus is used to publish and subscribe to events.
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
func (b *Bus) Publish(msg Message) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	subscribers := make(map[int]any)
	for id := range b.events[msg.Name] {
		subscribers[id] = struct{}{}
	}
	for id := range b.events[WildCardEventName] {
		subscribers[id] = struct{}{}
	}

	for id := range subscribers {
		select {
		case b.subs[id].value <- msg:
			// published event
		default:
			// channel full
		}
	}
}

// Subscribe returns a new channel that will receive all of the subscribed events.
func (b *Bus) Subscribe(size int, events ...string) *Subscription {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	sub := &Subscription{
		id:     b.subId,
		value:  make(chan Message, size),
		events: events,
	}

	for _, event := range events {
		if _, ok := b.events[event]; !ok {
			b.events[event] = make(map[int]any)
		}
		b.events[event][sub.id] = struct{}{}
	}

	b.subId++
	b.subs[sub.id] = sub
	return sub
}

// Unsubscribe unsubscribes from all events and closes the event channel of the given subscription.
func (b *Bus) Unsubscribe(sub *Subscription) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if _, ok := b.subs[sub.id]; !ok {
		return // not subscribed
	}
	for _, event := range sub.events {
		delete(b.events[event], sub.id)
	}

	delete(b.subs, sub.id)
	close(sub.value)
}

// Close closes the event bus by unsubscribing all subscribers.
func (b *Bus) Close() {
	var subs []*Subscription

	b.mutex.RLock()
	for _, sub := range b.subs {
		subs = append(subs, sub)
	}
	b.mutex.RUnlock()

	for _, sub := range subs {
		b.Unsubscribe(sub)
	}
}
