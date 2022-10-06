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

type Subscription[T any] struct {
	Channel chan T
	Closing chan struct{}
}

func newSubscription[T any](eventBufferSize int) Subscription[T] {
	return Subscription[T]{
		Channel: make(chan T, eventBufferSize),
		Closing: make(chan struct{}),
	}
}

func (sub *Subscription[T]) close() {
	close(sub.Closing)
	close(sub.Channel)
}
