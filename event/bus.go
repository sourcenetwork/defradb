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

// Bus is an event bus used to broadcasts messages to subscribers.
type Bus interface {
	// Subscribe subscribes to the Channel, returning a channel by which events can
	// be read from, or an error should one occur (e.g. if this object is closed).
	//
	// This function is non-blocking unless the subscription-buffer is full.
	Subscribe(events ...Name) (*Subscription, error)

	// Unsubscribe unsubscribes from the Channel, closing the provided channel.
	//
	// Will do nothing if this object is already closed.
	Unsubscribe(sub *Subscription)

	// Publish pushes the given item into this channel. Non-blocking.
	Publish(msg Message)

	// Close closes this Channel, and any owned or subscribing channels.
	Close()
}

// Message contains event info.
type Message struct {
	// Name is the name of the event this message was generated from.
	Name Name

	// Data contains optional event information.
	Data any
}

// NewMessage returns a new message with the given name and optional data.
func NewMessage(name Name, data any) Message {
	return Message{name, data}
}

// Subscription is a read-only event stream.
type Subscription struct {
	id     uint64
	value  chan Message
	events []Name
}

// Message returns the next event value from the subscription.
func (s *Subscription) Message() <-chan Message {
	return s.value
}

// Events returns the names of all subscribed events.
func (s *Subscription) Events() []Name {
	return s.events
}
