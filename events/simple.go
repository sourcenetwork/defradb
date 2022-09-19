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

import "github.com/sourcenetwork/defradb/errors"

type simpleEventChannel[T any] struct {
	subscribers         []chan T
	subscriptionChannel chan chan T
	unsubscribeChannel  chan chan T
	eventChannel        chan T
	eventBufferSize     int
	closeChannel        chan struct{}
	isClosed            bool
}

// NewSimpleEventChannel creates a new simpleEventChannel with the given subscriberBufferSize and
// eventBufferSize.
//
// Should the buffers be filled subsequent calls to functions on this object may start to block.
func NewSimpleEventChannel[T any](subscriberBufferSize int, eventBufferSize int) EventChannel[T] {
	c := simpleEventChannel[T]{
		subscriptionChannel: make(chan chan T, subscriberBufferSize),
		unsubscribeChannel:  make(chan chan T, subscriberBufferSize),
		eventChannel:        make(chan T, eventBufferSize),
		closeChannel:        make(chan struct{}),
	}

	go c.handleSelf()

	return &c
}

func (c *simpleEventChannel[T]) Subscribe() (chan T, error) {
	if c.isClosed {
		return nil, errors.New("cannot subscribe to a closed channel")
	}

	// It is important to set this buffer size too, else we may end up blocked in the handleSelf func
	ch := make(chan T, c.eventBufferSize)

	c.subscriptionChannel <- ch
	return ch, nil
}

func (c *simpleEventChannel[T]) Unsubscribe(ch chan T) {
	if c.isClosed {
		return
	}
	c.unsubscribeChannel <- ch
}

func (c *simpleEventChannel[T]) Push(item T) {
	if c.isClosed {
		return
	}
	c.eventChannel <- item
}

func (c *simpleEventChannel[T]) Close() {
	if c.isClosed {
		return
	}
	c.closeChannel <- struct{}{}
}

func (c *simpleEventChannel[T]) handleSelf() {
	for {
		select {
		case <-c.closeChannel:
			c.isClosed = true
			close(c.closeChannel)
			for _, subscriber := range c.subscribers {
				close(subscriber)
			}
			close(c.subscriptionChannel)
			close(c.unsubscribeChannel)
			close(c.eventChannel)
			return

		case ch := <-c.unsubscribeChannel:
			var isFound bool
			var index int
			for i, subscriber := range c.subscribers {
				if ch == subscriber {
					index = i
					isFound = true
				}
			}
			if !isFound {
				continue
			}

			// Remove channel from list of subscribers
			c.subscribers[index] = c.subscribers[len(c.subscribers)-1]
			c.subscribers = c.subscribers[:len(c.subscribers)-1]

			close(ch)

		case newSubscriber := <-c.subscriptionChannel:
			c.subscribers = append(c.subscribers, newSubscriber)

		case item := <-c.eventChannel:
			for _, subscriber := range c.subscribers {
				subscriber <- item
			}
		}
	}
}
