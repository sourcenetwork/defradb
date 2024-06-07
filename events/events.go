// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

/*
Package events provides the internal event system.
*/
package events

type Subscription[T any] chan T

// Channel represents a subscribable type that will expose inputted items to subscribers.
type Channel[T any] interface {
	// Subscribe subscribes to the Channel, returning a channel by which events can
	// be read from, or an error should one occur (e.g. if this object is closed).
	//
	// This function is non-blocking unless the subscription-buffer is full.
	Subscribe() (Subscription[T], error)

	// Unsubscribe unsubscribes from the Channel, closing the provided channel.
	//
	// Will do nothing if this object is already closed.
	Unsubscribe(Subscription[T])

	// Publish pushes the given item into this channel. Non-blocking.
	Publish(item T)

	// Close closes this Channel, and any owned or subscribing channels.
	Close()
}

var _ Channel[int] = (*simpleChannel[int])(nil)

// New creates and returns a new Channel instance.
//
// At the moment this will always return a new simpleChannel, however that may change in
// the future as this feature gets fleshed out.
func New[T any](commandBufferSize int, eventBufferSize int) Channel[T] {
	return NewSimpleChannel[T](commandBufferSize, eventBufferSize)
}

// Events hold the supported event types
type Events struct {
	// Updates publishes an `Update` for each document written to in the database.
	Updates UpdateChannel

	// DAGMerges publishes a `DAGMerge` for each completed DAG sync process over P2P.
	DAGMerges DAGMergeChannel
}
