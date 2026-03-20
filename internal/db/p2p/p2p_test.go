// Copyright 2026 Democratized Data Foundation
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
	"testing"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
		Documents: []protocol.DocumentInfo{{
			DocID: "docID",
		}},
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
		Documents: []protocol.DocumentInfo{{
			DocID: "docID",
		}},
	}
	msg, err := cbor.Marshal(req)
	assert.NoError(t, err)

	resp, err := p.pubSubMessageHandler("sender", "topic", msg)

	assert.NoError(t, err)
	assert.Nil(t, resp)
}

// --- processQueue tests ---

func TestProcessQueue_AddAndDone(t *testing.T) {
	pq := newProcessQueue()

	// add should not block for a new key
	done := make(chan struct{})
	go func() {
		pq.add("key1")
		close(done)
	}()

	select {
	case <-done:
		// good: add returned immediately
	case <-time.After(time.Second):
		t.Fatal("add blocked unexpectedly for a new key")
	}

	// key should be present in the map
	pq.mutex.Lock()
	_, exists := pq.keys["key1"]
	pq.mutex.Unlock()
	require.True(t, exists, "key should be in the queue after add")

	// calling done should remove the key
	pq.done("key1")

	pq.mutex.Lock()
	_, exists = pq.keys["key1"]
	pq.mutex.Unlock()
	require.False(t, exists, "key should be removed from the queue after done")
}

func TestProcessQueue_ConcurrentSameKey(t *testing.T) {
	pq := newProcessQueue()

	// Add the key from the first goroutine
	pq.add("key1")

	// A second goroutine tries to add the same key — it should block
	unblocked := make(chan struct{})
	go func() {
		pq.add("key1") // blocks until key1 is done
		close(unblocked)
	}()

	// Give the goroutine time to start and block
	time.Sleep(20 * time.Millisecond)

	select {
	case <-unblocked:
		t.Fatal("second add should have blocked but returned immediately")
	default:
		// good: second goroutine is still blocked
	}

	// Release the first hold
	pq.done("key1")

	// Now the second goroutine should unblock
	select {
	case <-unblocked:
		// good
	case <-time.After(time.Second):
		t.Fatal("second add did not unblock after done was called")
	}

	// Clean up
	pq.done("key1")
}

func TestProcessQueue_DoneOnce(t *testing.T) {
	pq := newProcessQueue()
	pq.add("key1")

	fn := pq.doneOnce("key1")

	// First call should succeed without panic
	require.NotPanics(t, fn, "first doneOnce call should not panic")

	// Second call should also not panic (sync.OnceFunc guarantees this)
	require.NotPanics(t, fn, "second doneOnce call should not panic")

	// The key should have been removed after the first call
	pq.mutex.Lock()
	_, exists := pq.keys["key1"]
	pq.mutex.Unlock()
	require.False(t, exists, "key should be removed after doneOnce")
}

// --- cachedCollection tests ---

func TestCachedCollection_Expiry(t *testing.T) {
	// Create a cachedCollection that is already expired
	expired := &cachedCollection{
		exists:  true,
		col:     nil,
		expires: time.Now().Add(-time.Second),
	}

	// Create a cachedCollection that has not yet expired
	fresh := &cachedCollection{
		exists:  true,
		col:     nil,
		expires: time.Now().Add(time.Hour),
	}

	p := &P2P{
		collectionCache: map[string]*cachedCollection{
			"expired": expired,
			"fresh":   fresh,
		},
	}

	// getCachedCollection for a non-expired entry should return the cached value
	result := p.getCachedCollection("fresh")
	// col is nil but the entry exists and is not expired; result matches col field
	assert.Nil(t, result, "nil col field should be returned as nil from getCachedCollection")

	// getCachedCollection for an expired entry should return nil
	result = p.getCachedCollection("expired")
	assert.Nil(t, result, "expired cache entry should return nil")

	// Verify that the TTL expiry boundary works: use a very short TTL
	shortLived := &cachedCollection{
		exists:  true,
		col:     nil,
		expires: time.Now().Add(10 * time.Millisecond),
	}
	p.collectionCache["short"] = shortLived

	// Before expiry: entry exists (col is nil but the cache hit is verified via the exists field)
	p.collectionCacheMu.RLock()
	cachedBefore, ok := p.collectionCache["short"]
	p.collectionCacheMu.RUnlock()
	require.True(t, ok)
	require.True(t, time.Now().Before(cachedBefore.expires), "entry should not yet be expired")

	// Wait for TTL to pass
	time.Sleep(20 * time.Millisecond)

	// After expiry: getCachedCollection should return nil
	result = p.getCachedCollection("short")
	assert.Nil(t, result, "cache entry should be nil after expiry")
}

// --- pubSubMessageHandler additional tests ---

func TestPubSubMessageHandler_QueueFull(t *testing.T) {
	ctx := context.Background()

	// Create a queue of capacity 1 and fill it
	msgQueue := make(chan *protocol.PushLogRequest, 1)
	filler := &protocol.PushLogRequest{CollectionID: "col1"}
	msgQueue <- filler // fill the only slot

	p := &P2P{
		ctx:      ctx,
		host:     &SimpleMockHost{},
		msgQueue: msgQueue,
	}

	req := protocol.PushLogRequest{
		CollectionID: "col1",
		Documents: []protocol.DocumentInfo{{
			DocID: "docID",
		}},
	}
	msg, err := cbor.Marshal(req)
	require.NoError(t, err)

	// pubSubMessageHandler should drop the message (queue full) and return nil error
	resp, err := p.pubSubMessageHandler("sender", "topic", msg)
	assert.NoError(t, err)
	assert.Nil(t, resp)

	// The queue should still only have the original filler message (the new one was dropped)
	assert.Equal(t, 1, len(msgQueue))
}

func TestPubSubMessageHandler_SelfMessage(t *testing.T) {
	ctx := context.Background()

	p := &P2P{
		ctx:      ctx,
		host:     &SimpleMockHost{}, // ID() returns "peerID"
		msgQueue: make(chan *protocol.PushLogRequest, 10),
	}

	req := protocol.PushLogRequest{
		CollectionID: "col1",
		Documents: []protocol.DocumentInfo{{
			DocID: "docID",
		}},
	}
	msg, err := cbor.Marshal(req)
	require.NoError(t, err)

	// Sending a message from self (from == host.ID()) should be ignored
	resp, err := p.pubSubMessageHandler("peerID", "topic", msg)
	assert.NoError(t, err)
	assert.Nil(t, resp)

	// Nothing should have been enqueued
	assert.Equal(t, 0, len(p.msgQueue))
}
