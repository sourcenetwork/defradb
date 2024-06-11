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

	"github.com/stretchr/testify/assert"
)

func TestBusPublish(t *testing.T) {
	bus := NewBus()
	defer bus.Close()

	sub1 := bus.Subscribe(1, "test")
	sub2 := bus.Subscribe(1, WildCardEventName)

	msg := NewMessage("test", "hello")
	bus.Publish(msg)

	event1 := <-sub1.Message()
	assert.Equal(t, msg, event1)

	event2 := <-sub2.Message()
	assert.Equal(t, msg, event2)
}
