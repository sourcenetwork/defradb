// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build js

package event

import (
	"slices"
	"sync"
	"sync/atomic"
	"syscall/js"

	"github.com/sourcenetwork/goji"
)

type eventTargetSub struct {
	id       uint64
	listener js.Func
	value    chan Message
	events   []Name
}

// Message returns the message channel for the subscription.
func (s *eventTargetSub) Message() <-chan Message {
	return s.value
}

// eventTargetBus uses a JavaScript EventTarget to
// manage subscribers and publish messages.
type eventTargetBus struct {
	target       goji.EventTargetValue
	bufferSize   int
	subscriberID atomic.Uint64
	subscribers  map[uint64]*eventTargetSub
	closeMutex   sync.Mutex
}

// NewEventTargetBus creates a new bus from the given JavaScript EventTarget.
//
// This bus is meant for use in JavaScript enviroments (browser and NodeJS).
//
// Messages are serialized to JSON before being unserialized to a js.Value.
func NewEventTargetBus(value js.Value, bufferSize int) Bus {
	return &eventTargetBus{
		target:      goji.EventTargetValue(value),
		bufferSize:  bufferSize,
		subscribers: make(map[uint64]*eventTargetSub),
	}
}

// Publish broadcasts the given message to the bus subscribers. Non-blocking.
func (b *eventTargetBus) Publish(msg Message) {
	detail := goji.MustMarshalJS(msg.Data)
	event := goji.CustomEvent.New(string(msg.Name), detail)
	b.target.DispatchEvent(js.Value(event))
}

// Subscribe returns a new subscription that will receive all of the events
// contained in the given list of events.
func (b *eventTargetBus) Subscribe(events ...Name) (Subscription, error) {
	b.closeMutex.Lock()
	defer b.closeMutex.Unlock()

	if slices.Contains(events, WildCardName) {
		return nil, ErrWildcardNotSupported
	}

	value := make(chan Message, b.bufferSize)
	listener := goji.EventListener(func(event goji.EventValue) {
		value <- unmarshalMessage(event)
	})

	sub := &eventTargetSub{
		id:       b.subscriberID.Add(1),
		listener: listener,
		value:    value,
		events:   events,
	}
	b.subscribers[sub.id] = sub

	for _, e := range events {
		b.target.AddEventListener(string(e), listener.Value)
	}
	return sub, nil
}

// Unsubscribe removes all event subscriptions and closes the subscription.
//
// Will do nothing if this object is already closed.
func (b *eventTargetBus) Unsubscribe(sub Subscription) {
	b.closeMutex.Lock()
	defer b.closeMutex.Unlock()

	s, ok := sub.(*eventTargetSub)
	if !ok {
		panic("failed to unsubscribe: invalid subscription type")
	}
	b.unsubscribe(s)
}

func (b *eventTargetBus) unsubscribe(sub *eventTargetSub) {
	if _, ok := b.subscribers[sub.id]; !ok {
		return
	}
	for _, e := range sub.events {
		b.target.RemoveEventListener(string(e), sub.listener.Value, js.Undefined())
	}
	close(sub.value)
	delete(b.subscribers, sub.id)
	sub.listener.Release()
}

// Close unsubscribes all active subscribers and closes the bus.
func (b *eventTargetBus) Close() {
	b.closeMutex.Lock()
	defer b.closeMutex.Unlock()

	for _, s := range b.subscribers {
		b.unsubscribe(s)
	}
}

// unmarshalMessage unmarshals a message from a JS EventValue.
//
// If the message type is unknown the data will be unmarshalled to a basic go type.
func unmarshalMessage(event goji.EventValue) Message {
	message := Message{
		Name: Name(event.Type()),
	}
	detail := js.Value(event).Get("detail")
	if detail.IsUndefined() {
		return message
	}
	switch message.Name {
	case MergeName:
		var value Merge
		goji.MustUnmarshalJS(detail, &value)
		message.Data = value
	case MergeCompleteName:
		var value MergeComplete
		goji.MustUnmarshalJS(detail, &value)
		message.Data = value
	case UpdateName:
		var value Update
		goji.MustUnmarshalJS(detail, &value)
		message.Data = value
	case PubSubName:
		var value PubSub
		goji.MustUnmarshalJS(detail, &value)
		message.Data = value
	case PeerInfoName:
		var value PeerInfo
		goji.MustUnmarshalJS(detail, &value)
		message.Data = value
	case ReplicatorName:
		var value Replicator
		goji.MustUnmarshalJS(detail, &value)
		message.Data = value
	case ReplicatorFailureName:
		var value ReplicatorFailure
		goji.MustUnmarshalJS(detail, &value)
		message.Data = value
	default:
		goji.MustUnmarshalJS(detail, &message.Data)
	}
	return message
}
