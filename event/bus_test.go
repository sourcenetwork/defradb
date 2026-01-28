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
	"errors"
	"sync"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type BusTestSuite struct {
	suite.Suite
	setup func(int) Bus
}

func NewBusTestSuite(setup func(int) Bus) *BusTestSuite {
	return &BusTestSuite{
		setup: setup,
	}
}

func (b *BusTestSuite) TestBus_IfPublishingWithoutSubscribers_ItShouldNotBlock() {
	bus := b.setup(0)
	defer bus.Close()

	msg := NewMessage("test", 1)
	bus.Publish(msg)

	// just assert that we reach this line, for the sake of having an assert
	assert.True(b.T(), true)
}

func (b *BusTestSuite) TestBus_IfClosingAfterSubscribing_ItShouldNotBlock() {
	bus := b.setup(0)
	defer bus.Close()

	sub, err := bus.Subscribe("test")
	b.Assert().NoError(err)

	bus.Close()

	<-sub.Message()

	// just assert that we reach this line, for the sake of having an assert
	b.Assert().True(true)
}

func (b *BusTestSuite) TestBus_IfSubscriptionIsUnsubscribedTwice_ItShouldNotPanic() {
	bus := b.setup(0)
	defer bus.Close()

	sub, err := bus.Subscribe("test")
	b.Assert().NoError(err)

	bus.Unsubscribe(sub)
	bus.Unsubscribe(sub)
}

func (b *BusTestSuite) TestBus_IfSubscribedToWildCard_ItShouldNotReceiveMessageTwice() {
	bus := b.setup(0)
	defer bus.Close()

	sub, err := bus.Subscribe("test", WildCardName)
	if errors.Is(err, ErrWildcardNotSupported) {
		b.T().Skipf("wildcard not supported")
	}
	b.Assert().NoError(err)

	msg := NewMessage("test", 1)
	bus.Publish(msg)

	evt := <-sub.Message()
	b.Assert().Equal(evt, msg)

	select {
	case <-sub.Message():
		b.T().Errorf("should not receive duplicate message")
	case <-time.After(100 * time.Millisecond):
		// message is deduplicated
	}
}

func (b *BusTestSuite) TestBus_IfMultipleSubscriptionsToTheSameEvent_EachSubscriberRecievesEachEvent() {
	bus := b.setup(0)
	defer bus.Close()

	msg1 := NewMessage("test", false)
	msg2 := NewMessage("test", true)

	sub1, err := bus.Subscribe("test")
	b.Assert().NoError(err)

	sub2, err := bus.Subscribe("test")
	b.Assert().NoError(err)

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

	b.Assert().Equal(msg1, event1)
	b.Assert().Equal(msg1, event2)

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

	b.Assert().Equal(msg2, event1)
	b.Assert().Equal(msg2, event2)
}

func (b *BusTestSuite) TestBus_IfMultipleBufferedSubscribersWithMultipleEvents_EachSubscriberRecievesEachItem() {
	bus := b.setup(2)
	defer bus.Close()

	msg1 := NewMessage("test", false)
	msg2 := NewMessage("test", true)

	sub1, err := bus.Subscribe("test")
	b.Assert().NoError(err)
	sub2, err := bus.Subscribe("test")
	b.Assert().NoError(err)

	// both inputs are added first before read, using the internal chan buffer
	bus.Publish(msg1)
	bus.Publish(msg2)

	output1Ch1 := <-sub1.Message()
	output1Ch2 := <-sub2.Message()

	output2Ch1 := <-sub1.Message()
	output2Ch2 := <-sub2.Message()

	b.Assert().Equal(msg1, output1Ch1)
	b.Assert().Equal(msg1, output1Ch2)

	b.Assert().Equal(msg2, output2Ch1)
	b.Assert().Equal(msg2, output2Ch2)
}

func (b *BusTestSuite) TestBus_IfSubscribedThenUnsubscribe_SubscriptionShouldNotReceiveEvent() {
	bus := b.setup(0)
	defer bus.Close()

	sub, err := bus.Subscribe("test")
	b.Assert().NoError(err)
	bus.Unsubscribe(sub)

	msg := NewMessage("test", 1)
	bus.Publish(msg)

	// tiny delay to try and make sure the internal logic would have had time
	// to do its thing with the pushed item.
	time.Sleep(5 * time.Millisecond)

	// closing the channel will result in reads yielding the default value
	b.Assert().Equal(Message{}, <-sub.Message())
}
