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

type simpleChannel[T any] struct {
	subscribers         []Subscription[T]
	subscriptionChannel chan Subscription[T]
	unsubscribeChannel  chan Subscription[T]
	eventChannel        chan T
	eventBufferSize     int
	closeChannel        chan struct{}
	isClosed            bool
}

// NewSimpleChannel creates a new simpleChannel with the given subscriberBufferSize and
// eventBufferSize.
//
// Should the buffers be filled subsequent calls to functions on this object may start to block.
func NewSimpleChannel[T any](subscriberBufferSize int, eventBufferSize int) Channel[T] {
	c := simpleChannel[T]{
		subscriptionChannel: make(chan Subscription[T], subscriberBufferSize),
		unsubscribeChannel:  make(chan Subscription[T], subscriberBufferSize),
		eventChannel:        make(chan T, eventBufferSize),
		eventBufferSize:     eventBufferSize,
		closeChannel:        make(chan struct{}),
	}

	go c.handleChannel()

	return &c
}

func (c *simpleChannel[T]) Subscribe() (Subscription[T], error) {
	if c.isClosed {
		return Subscription[T]{}, errors.New("cannot subscribe to a closed channel")
	}

	// It is important to set this buffer size too, else we may end up blocked in the handleChannel func
	ch := newSubscription[T](c.eventBufferSize)

	c.subscriptionChannel <- ch
	return ch, nil
}

func (c *simpleChannel[T]) Unsubscribe(ch Subscription[T]) {
	if c.isClosed {
		return
	}
	c.unsubscribeChannel <- ch
}

func (c *simpleChannel[T]) Publish(item T) {
	if c.isClosed {
		return
	}
	c.eventChannel <- item
}

func (c *simpleChannel[T]) Close() {
	if c.isClosed {
		return
	}
	c.isClosed = true
	c.closeChannel <- struct{}{}
}

func (c *simpleChannel[T]) handleChannel() {
	for {
		select {
		case <-c.closeChannel:
			close(c.closeChannel)
			for _, subscriber := range c.subscribers {
				subscriber.close()
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
					break
				}
			}
			if !isFound {
				continue
			}

			// Remove channel from list of subscribers
			c.subscribers[index] = c.subscribers[len(c.subscribers)-1]
			c.subscribers = c.subscribers[:len(c.subscribers)-1]

			ch.close()

		case newSubscriber := <-c.subscriptionChannel:
			c.subscribers = append(c.subscribers, newSubscriber)

		case item := <-c.eventChannel:
			for _, subscriber := range c.subscribers {
				subscriber.Channel <- item
			}
		}
	}
}
