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

type EventChannel[T any] interface {
	// Subscribe subscribes to the EventChannel, returning a channel by which events can
	// be read from, or an error should one occur (e.g. if this object is closed).
	//
	// This function is non-blocking unless the subscription-buffer is full.
	Subscribe() (chan T, error)

	// Unsubscribe unsubscribes from the EventChannel, closing the provided channel.
	//
	// Will do nothing if this object is already closed.
	Unsubscribe(chan T)

	// Push pushes the given item into this channel. Non-blocking.
	Push(item T)

	// Close closes this EventChannel, and any owned or subscribing channels.
	Close()
}

var _ EventChannel[int] = (*simpleEventChannel[int])(nil)

// New creates and returns a new EventChannel instance.
//
// At the moment this will always return a new simpleEventChannel, however that may change in
// the future as this feature gets fleshed out.
func New[T any](subscriberBufferSize int, eventBufferSize int) EventChannel[T] {
	return NewSimpleEventChannel[T](subscriberBufferSize, eventBufferSize)
}
