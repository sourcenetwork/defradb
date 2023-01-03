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

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewPublisher(t *testing.T) {
	ch := startEventChanel()

	pub, err := NewPublisher(ch, 0)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, pub)
}

func TestNewPublisherWithError(t *testing.T) {
	ch := startEventChanel()
	ch.Close()
	_, err := NewPublisher(ch, 0)
	assert.Error(t, err)
}

func TestPublisherToStream(t *testing.T) {
	ch := startEventChanel()

	pub, err := NewPublisher(ch, 1)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, pub)

	ch.Publish(10)
	evt := <-pub.Event()
	assert.Equal(t, 10, evt)

	pub.Publish(evt)
	assert.Equal(t, 10, <-pub.Stream())

	pub.Unsubscribe()

	_, open := <-pub.Stream()
	assert.Equal(t, false, open)
}

func TestPublisherToStreamWithTimeout(t *testing.T) {
	clientTimeout = 1 * time.Second
	ch := startEventChanel()

	pub, err := NewPublisher(ch, 0)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, pub)

	ch.Publish(10)
	evt := <-pub.Event()
	assert.Equal(t, 10, evt)

	pub.Publish(evt)

	_, open := <-pub.Stream()
	assert.Equal(t, false, open)
}

func startEventChanel() Channel[int] {
	return New[int](0, 0)
}
