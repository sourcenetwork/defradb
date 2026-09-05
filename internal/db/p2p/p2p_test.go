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
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/db/p2p/protocol"
)

type SimpleMockHost struct {
	client.Host
}

// withReasonMaps gives p the counters New builds, so a test can construct a P2P from a literal
// and still record.
func withReasonMaps(p *P2P) *P2P {
	p.initReasonCounters()
	return p
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

// The handler reserves the budget and the worker is what gives it back. Without the release the
// counter only ever climbs, and once enough traffic has passed through, the node drops every
// message from then on while its queue sits empty.
func TestProcessMessageWorker_ReleasesByteBudget(t *testing.T) {
	msg, err := cbor.Marshal(protocol.PushLogRequest{DocID: "docID"})
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	p := &P2P{
		ctx:              ctx,
		host:             &SimpleMockHost{},
		msgQueue:         make(chan queuedMessage, 1),
		msgQueueMaxBytes: int64(len(msg)),
	}

	_, err = p.pubSubMessageHandler("sender", "topic", msg)
	assert.NoError(t, err)
	assert.Equal(t, int64(len(msg)), p.msgQueueBytes.Load(), "the handler reserves on enqueue")

	// Cancelling returns processPushlogRequest at its first check, so no database is needed and
	// the release is the only part of the worker left under test. The worker is left parked on
	// the empty queue rather than stopped, which keeps this independent of how shutdown is
	// signalled.
	cancel()
	p.msgWorkers.Add(1)
	go p.processMessageWorker()

	assert.Eventually(t, func() bool { return p.msgQueueBytes.Load() == 0 },
		5*time.Second, 5*time.Millisecond,
		"the worker must return the budget the handler reserved")
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

	// Written out rather than derived from the constants: the point of the assertion is that
	// the budget is a small fraction of the limit, and restating the formula would let the
	// queue grow to the whole limit without failing here.
	debug.SetMemoryLimit(8 << 30)
	assert.Equal(t, int64(2<<30), queueByteBudget(), "a quarter of an 8 GiB limit")

	debug.SetMemoryLimit(math.MaxInt64)
	assert.Equal(t, int64(1<<30), queueByteBudget(), "the fallback when no limit is set")

	// Zero is a limit an operator can set, so reading it as "no limit" would give the
	// tightest heap the largest budget.
	debug.SetMemoryLimit(0)
	assert.Equal(t, int64(msgQueueMinBytes), queueByteBudget(), "a zero limit is floored, not ignored")

	debug.SetMemoryLimit(1 << 10)
	assert.Equal(t, int64(msgQueueMinBytes), queueByteBudget(), "a limit too small for one message is floored")
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

// denseBatch builds a batch of empty entries, the smallest encoding that still allocates a
// DocumentInfo each. Hand-built because the encoder writes every field without omitempty and
// so cannot produce an encoding a peer is free to send.
func denseBatch(n int) []byte {
	out := []byte{0xa1, 0x69} // map of one pair, keyed by a nine byte text string
	out = append(out, []byte("documents")...)
	out = append(out, 0x9a, byte(n>>24), byte(n>>16), byte(n>>8), byte(n)) // array, four byte length
	for range n {
		out = append(out, 0xa0) // empty map
	}
	return out
}

// A budget taken from the wire size alone admits a message holding far more than it was
// charged for, since an entry costs a struct however little it carries.
func TestPubSubMessageHandler_ChargesForEachDocument(t *testing.T) {
	msg := denseBatch(1024)

	p := &P2P{
		ctx:      context.Background(),
		host:     &SimpleMockHost{},
		msgQueue: make(chan queuedMessage, 4),
		// Room for the wire bytes several times over, but not for what the batch holds.
		msgQueueMaxBytes: int64(len(msg)) * 4,
	}

	_, err := p.pubSubMessageHandler("sender", "topic", msg)
	assert.NoError(t, err)
	assert.Empty(t, p.msgQueue, "a batch holding more than the budget must not be queued")
	assert.Equal(t, int64(0), p.msgQueueBytes.Load(), "a rejected message must hold no budget")
}

// Every path that gives up on a message must give back what was claimed for it, or the
// counter climbs until a node with an empty queue drops everything.
func TestPubSubMessageHandler_QueueFullReleasesTheWholeReservation(t *testing.T) {
	msg := denseBatch(64)

	p := &P2P{
		ctx:              context.Background(),
		host:             &SimpleMockHost{},
		msgQueue:         make(chan queuedMessage, 1),
		msgQueueMaxBytes: 1 << 30,
	}

	_, err := p.pubSubMessageHandler("sender", "topic", msg)
	assert.NoError(t, err)
	assert.Len(t, p.msgQueue, 1)
	held := p.msgQueueBytes.Load()

	for range 8 {
		_, err = p.pubSubMessageHandler("sender", "topic", msg)
		assert.NoError(t, err)
	}
	assert.Equal(t, held, p.msgQueueBytes.Load(), "a full queue must give back everything it claimed")

	queued := <-p.msgQueue
	p.releaseQueueBytes(queued.size)
	assert.Equal(t, int64(0), p.msgQueueBytes.Load(), "the reservation must balance to zero")
}

// The cap belongs at the decoder: a length check after it runs past the allocation it exists
// to prevent.
func TestPubSubMessageHandler_RejectsAnOversizedBatch(t *testing.T) {
	p := &P2P{
		ctx:              context.Background(),
		host:             &SimpleMockHost{},
		msgQueue:         make(chan queuedMessage, 4),
		msgQueueMaxBytes: 1 << 30,
	}

	_, err := p.pubSubMessageHandler("sender", "topic", denseBatch(maxInboundDocuments+1))
	assert.Error(t, err, "a batch over the entry cap must not decode")
	assert.Empty(t, p.msgQueue)
	assert.Equal(t, int64(0), p.msgQueueBytes.Load(), "a rejected message must hold no budget")

	_, err = p.pubSubMessageHandler("sender", "topic", denseBatch(maxInboundDocuments))
	assert.NoError(t, err)
	assert.Len(t, p.msgQueue, 1, "a batch at the cap must still be accepted")
}
