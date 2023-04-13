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

type simpleChannel[T any] struct {
	subscribers []chan T
	// commandChannel manages all commands sent to this simpleChannel.
	//
	// It is important that all stuff gets sent through this single channel to ensure
	// that the order of operations is preserved.
	commandChannel  chan any
	eventBufferSize int
	hasClosedChan   chan struct{}
	isClosed        bool
}

type subscribeCommand[T any] struct {
	subscriptionChannel Subscription[T]
}

type unsubscribeCommand[T any] struct {
	subscriptionChannel Subscription[T]
}

type publishCommand[T any] struct {
	item T
}

type closeCommand struct{}

// NewSimpleChannel creates a new simpleChannel with the given commandBufferSize and
// eventBufferSize.
//
// Should the buffers be filled subsequent calls to functions on this object may start to block.
func NewSimpleChannel[T any](commandBufferSize int, eventBufferSize int) Channel[T] {
	c := simpleChannel[T]{
		commandChannel:  make(chan any, commandBufferSize),
		hasClosedChan:   make(chan struct{}),
		eventBufferSize: eventBufferSize,
	}

	go c.handleChannel()

	return &c
}

func (c *simpleChannel[T]) Subscribe() (Subscription[T], error) {
	if c.isClosed {
		return nil, ErrSubscribedToClosedChan
	}

	// It is important to set this buffer size too, else we may end up blocked in the handleChannel func
	ch := make(chan T, c.eventBufferSize)

	c.commandChannel <- subscribeCommand[T]{ch}
	return ch, nil
}

func (c *simpleChannel[T]) Unsubscribe(ch Subscription[T]) {
	if c.isClosed {
		return
	}
	c.commandChannel <- unsubscribeCommand[T]{ch}
}

func (c *simpleChannel[T]) Publish(item T) {
	if c.isClosed {
		return
	}
	c.commandChannel <- publishCommand[T]{item}
}

func (c *simpleChannel[T]) Close() {
	if c.isClosed {
		return
	}
	c.isClosed = true
	c.commandChannel <- closeCommand{}

	// Wait for the close command to be handled, in order, before returning
	<-c.hasClosedChan
}

func (c *simpleChannel[T]) handleChannel() {
	for cmd := range c.commandChannel {
		switch command := cmd.(type) {
		case closeCommand:
			for _, subscriber := range c.subscribers {
				close(subscriber)
			}
			close(c.commandChannel)
			close(c.hasClosedChan)
			return

		case subscribeCommand[T]:
			c.subscribers = append(c.subscribers, command.subscriptionChannel)

		case unsubscribeCommand[T]:
			var isFound bool
			var index int
			for i, subscriber := range c.subscribers {
				if command.subscriptionChannel == subscriber {
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

			close(command.subscriptionChannel)

		case publishCommand[T]:
			for _, subscriber := range c.subscribers {
				subscriber <- command.item
			}
		}
	}
}
