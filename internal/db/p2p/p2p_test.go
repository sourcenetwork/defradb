// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package p2p

import (
	"context"
	"math"
	"runtime/debug"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/db/p2p/protocol"
)

type SimpleMockHost struct {
	client.Host
}

func (m *SimpleMockHost) ID() string {
	return "peerID"
}

func TestPubSubMessageHandler_ContextCanceled(t *testing.T) {
	// Setup P2P with canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	p := &P2P{
		ctx:  ctx, // This should trigger the early exit in processPushlogRequest
		host: &SimpleMockHost{},
	}

	// Create a dummy request message
	req := protocol.PushLogRequest{
		DocID: "docID",
		// Block can be empty or garbage, as context check should happen first
	}
	msg, err := cbor.Marshal(req)
	assert.NoError(t, err)

	// Call handler
	// from="sender", topic="topic"
	resp, err := p.pubSubMessageHandler("sender", "topic", msg)

	// Expectation: No error returned (suppressed), resp is nil
	assert.NoError(t, err)
	assert.Nil(t, resp)
}

// The budget is claimed before the decode, so an over-budget message never reaches
// cbor. Undecodable bytes distinguish the two orderings: decoding first surfaces an
// error, claiming first drops the message silently.
func TestPubSubMessageHandler_OverBudgetDropsBeforeDecode(t *testing.T) {
	p := &P2P{
		ctx:              context.Background(),
		host:             &SimpleMockHost{},
		msgQueue:         make(chan queuedMessage, 4),
		msgQueueMaxBytes: 4,
	}

	undecodable := []byte{0xff, 0xff, 0xff, 0xff, 0xff}
	resp, err := p.pubSubMessageHandler("sender", "topic", undecodable)

	assert.NoError(t, err)
	assert.Nil(t, resp)
	assert.Empty(t, p.msgQueue)
	assert.Equal(t, int64(0), p.msgQueueBytes.Load())
	assert.Equal(t, int64(1), p.statDroppedBudget.Load())
	assert.Equal(t, int64(0), p.statDroppedFull.Load())
}

func TestPubSubMessageHandler_ByteBudgetReleasedAfterProcessing(t *testing.T) {
	msg, err := cbor.Marshal(protocol.PushLogRequest{DocID: "docID"})
	assert.NoError(t, err)

	// Sized so exactly one message fits, making the second one's fate depend entirely
	// on whether the first was released.
	p := &P2P{
		ctx:              context.Background(),
		host:             &SimpleMockHost{},
		msgQueue:         make(chan queuedMessage, 4),
		msgQueueMaxBytes: int64(len(msg)),
	}

	_, err = p.pubSubMessageHandler("sender", "topic", msg)
	assert.NoError(t, err)
	assert.Len(t, p.msgQueue, 1)

	_, err = p.pubSubMessageHandler("sender", "topic", msg)
	assert.NoError(t, err)
	assert.Len(t, p.msgQueue, 1, "second message should be dropped while the budget is held")

	// Stand in for a worker finishing with the message.
	queued := <-p.msgQueue
	p.releaseQueueBytes(queued.size)
	assert.Equal(t, int64(0), p.msgQueueBytes.Load())

	_, err = p.pubSubMessageHandler("sender", "topic", msg)
	assert.NoError(t, err)
	assert.Len(t, p.msgQueue, 1, "budget should be available again once released")
}

// A full queue and an exhausted budget are separate drop reasons, and the stats line is
// only useful if a drop is attributed to the one that caused it.
func TestPubSubMessageHandler_QueueFullCountedSeparately(t *testing.T) {
	msg, err := cbor.Marshal(protocol.PushLogRequest{DocID: "docID"})
	assert.NoError(t, err)

	// Budget large enough to stay out of the way, so only the slot count can reject.
	p := &P2P{
		ctx:              context.Background(),
		host:             &SimpleMockHost{},
		msgQueue:         make(chan queuedMessage, 1),
		msgQueueMaxBytes: 1 << 20,
	}

	_, err = p.pubSubMessageHandler("sender", "topic", msg)
	assert.NoError(t, err)
	_, err = p.pubSubMessageHandler("sender", "topic", msg)
	assert.NoError(t, err)

	assert.Equal(t, int64(1), p.statDroppedFull.Load())
	assert.Equal(t, int64(0), p.statDroppedBudget.Load())
	// The dropped message must not keep holding its reservation.
	assert.Equal(t, int64(len(msg)), p.msgQueueBytes.Load())
}

func TestQueueByteBudget_DerivesFromMemoryLimit(t *testing.T) {
	original := debug.SetMemoryLimit(-1)
	t.Cleanup(func() { debug.SetMemoryLimit(original) })

	debug.SetMemoryLimit(8 << 30)
	assert.Equal(t, int64(8<<30)/msgQueueMemDivisor, queueByteBudget())

	debug.SetMemoryLimit(math.MaxInt64)
	assert.Equal(t, int64(msgQueueFallbackMaxBytes), queueByteBudget())
}

func TestPubSubMessageHandler_ContextTimeout(t *testing.T) {
	// Setup P2P with timed out context
	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()

	// Wait for context to be done
	<-ctx.Done()

	p := &P2P{
		ctx:  ctx,
		host: &SimpleMockHost{},
	}

	req := protocol.PushLogRequest{
		DocID: "docID",
	}
	msg, err := cbor.Marshal(req)
	assert.NoError(t, err)

	resp, err := p.pubSubMessageHandler("sender", "topic", msg)

	assert.NoError(t, err)
	assert.Nil(t, resp)
}
