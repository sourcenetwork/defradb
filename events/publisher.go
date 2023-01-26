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

import "time"

// time limit we set for the client to read after publishing.
var clientTimeout = 60 * time.Second

// Publisher hold a referance to the event channel,
// the associated subscription channel and the stream channel that
// returns data to the subscribed client
type Publisher[T any] struct {
	ch     Channel[T]
	event  Subscription[T]
	stream chan any
}

// NewPublisher creates a new Publisher with the given event Channel, subscribes to the
// event Channel and opens a new channel for the stream.
func NewPublisher[T any](ch Channel[T], streamBufferSize int) (*Publisher[T], error) {
	evtCh, err := ch.Subscribe()
	if err != nil {
		return nil, err
	}

	return &Publisher[T]{
		ch:     ch,
		event:  evtCh,
		stream: make(chan any, streamBufferSize),
	}, nil
}

// Event returns the subscription channel
func (p *Publisher[T]) Event() Subscription[T] {
	return p.event
}

// Stream returns the streaming channel
func (p *Publisher[T]) Stream() chan any {
	return p.stream
}

// Publish sends data to the streaming channel and unsubscribes if
// the client hangs for too long.
func (p *Publisher[T]) Publish(data any) {
	select {
	case p.stream <- data:
	case <-time.After(clientTimeout):
		// if sending to the client times out, we assume an inactive or problematic client and
		// unsubscribe them from the event stream
		p.Unsubscribe()
	}
}

// Unsubscribe unsubscribes the client for the event channel and closes the stream.
func (p *Publisher[T]) Unsubscribe() {
	p.ch.Unsubscribe(p.event)
	close(p.stream)
}
