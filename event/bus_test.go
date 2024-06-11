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

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBusSubscribeThenPublish(t *testing.T) {
	bus := NewBus(100 * time.Millisecond)
	defer bus.Close()

	sub1 := bus.Subscribe(1, "test")
	sub2 := bus.Subscribe(1, WildCardEventName, "test")

	assert.ElementsMatch(t, sub1.Events(), []string{"test"})
	assert.ElementsMatch(t, sub2.Events(), []string{WildCardEventName, "test"})

	msg := NewMessage("test", "hello")
	go bus.Publish(msg)

	event := <-sub1.Message()
	assert.Equal(t, msg, event)

	event = <-sub2.Message()
	assert.Equal(t, msg, event)

	select {
	case <-sub2.Message():
		t.Fatalf("subscriber should not recieve duplicate message")
	case <-time.After(150 * time.Millisecond):
		// wait for publish timeout + skew
	}
}

func TestBusPublishThenSubscribe(t *testing.T) {
	bus := NewBus(100 * time.Millisecond)
	defer bus.Close()

	msg := NewMessage("test", "hello")
	bus.Publish(msg)

	sub := bus.Subscribe(1, "test")
	select {
	case <-sub.Message():
		t.Fatalf("subscriber should not recieve message")
	case <-time.After(150 * time.Millisecond):
		// wait for publish timeout + skew
	}
}

func TestBusSubscribeThenUnsubscribeThenPublish(t *testing.T) {
	bus := NewBus(100 * time.Millisecond)
	defer bus.Close()

	sub := bus.Subscribe(1, "test")
	bus.Unsubscribe(sub)

	msg := NewMessage("test", "hello")
	bus.Publish(msg)

	_, ok := <-sub.Message()
	assert.False(t, ok, "channel should be closed")
}

func TestBusUnsubscribeTwice(t *testing.T) {
	bus := NewBus(100 * time.Millisecond)
	defer bus.Close()

	sub := bus.Subscribe(1, "test")
	bus.Unsubscribe(sub)
	bus.Unsubscribe(sub)
}
