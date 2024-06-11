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
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSimplePushIsNotBlockedWithoutSubscribers(t *testing.T) {
	bus := NewBus(0, 0)
	defer bus.Close()

	msg := NewMessage("test", 1)
	bus.Publish(msg)

	// just assert that we reach this line, for the sake of having an assert
	assert.True(t, true)
}

func TestSimpleSubscribersAreNotBlockedAfterClose(t *testing.T) {
	bus := NewBus(0, 0)
	defer bus.Close()

	sub, err := bus.Subscribe("test")
	assert.Nil(t, err)

	bus.Close()

	<-sub.Message()

	// just assert that we reach this line, for the sake of having an assert
	assert.True(t, true)
}

func TestSimpleEachSubscribersRecievesEachItem(t *testing.T) {
	bus := NewBus(0, 0)
	defer bus.Close()

	msg1 := NewMessage("test", 1)
	msg2 := NewMessage("test", 2)

	sub1, err := bus.Subscribe("test")
	assert.Nil(t, err)

	sub2, err := bus.Subscribe("test")
	assert.Nil(t, err)

	// ordering of publish is not deterministic
	// so capture each in a go routine
	var wg sync.WaitGroup
	var event1 Message
	var event2 Message

	go func() {
		event1 = <-sub1.Message()
		wg.Done()
	}()

	go func() {
		event2 = <-sub2.Message()
		wg.Done()
	}()

	wg.Add(2)
	bus.Publish(msg1)
	wg.Wait()

	assert.Equal(t, msg1, event1)
	assert.Equal(t, msg1, event2)

	go func() {
		event1 = <-sub1.Message()
		wg.Done()
	}()

	go func() {
		event2 = <-sub2.Message()
		wg.Done()
	}()

	wg.Add(2)
	bus.Publish(msg2)
	wg.Wait()

	assert.Equal(t, msg2, event1)
	assert.Equal(t, msg2, event2)
}

func TestSimpleEachSubscribersRecievesEachItemGivenBufferedEventChan(t *testing.T) {
	bus := NewBus(0, 2)
	defer bus.Close()

	msg1 := NewMessage("test", 1)
	msg2 := NewMessage("test", 2)

	sub1, err := bus.Subscribe("test")
	assert.Nil(t, err)
	sub2, err := bus.Subscribe("test")
	assert.Nil(t, err)

	// both inputs are added first before read, using the internal chan buffer
	bus.Publish(msg1)
	bus.Publish(msg2)

	output1Ch1 := <-sub1.Message()
	output1Ch2 := <-sub2.Message()

	output2Ch1 := <-sub1.Message()
	output2Ch2 := <-sub2.Message()

	assert.Equal(t, msg1, output1Ch1)
	assert.Equal(t, msg1, output1Ch2)

	assert.Equal(t, msg2, output2Ch1)
	assert.Equal(t, msg2, output2Ch2)
}

func TestSimpleSubscribersDontRecieveItemsAfterUnsubscribing(t *testing.T) {
	bus := NewBus(0, 0)
	defer bus.Close()

	sub, err := bus.Subscribe("test")
	assert.Nil(t, err)
	bus.Unsubscribe(sub)

	msg := NewMessage("test", 1)
	bus.Publish(msg)

	// tiny delay to try and make sure the internal logic would have had time
	// to do its thing with the pushed item.
	time.Sleep(5 * time.Millisecond)

	// closing the channel will result in reads yielding the default value
	assert.Equal(t, Message{}, <-sub.Message())
}
