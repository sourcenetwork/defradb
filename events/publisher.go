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
const clientTimeout = 60 * time.Second

type Publisher[T any] struct {
	ch     Channel[T]
	event  Subscription[T]
	stream chan any
}

func NewPublisher[T any](ch Channel[T]) (*Publisher[T], error) {
	evtCh, err := ch.Subscribe()
	if err != nil {
		return nil, err
	}

	return &Publisher[T]{
		ch:     ch,
		event:  evtCh,
		stream: make(chan any),
	}, nil
}

func (p *Publisher[T]) Event() Subscription[T] {
	return p.event
}

func (p *Publisher[T]) Stream() chan any {
	return p.stream
}

func (p *Publisher[T]) Publish(data any) {
	select {
	case p.stream <- data:
	case <-time.After(clientTimeout):
		// if sending to the client times out, we assume an inactive or problematic client and
		// unsubscribe them from the event stream
		p.Unsubscribe()
	}
}

func (p *Publisher[T]) Unsubscribe() {
	p.ch.Unsubscribe(p.event)
	close(p.stream)
}
