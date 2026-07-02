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
	"sync"
	"time"

	"github.com/fxamacker/cbor/v2"

	"github.com/sourcenetwork/defradb/internal/db/p2p/protocol"
)

const (
	// batchFlushInterval is the maximum time to wait before flushing a pending batch.
	batchFlushInterval = 50 * time.Millisecond
	// batchMaxDocs is the maximum number of documents per batch before an immediate flush.
	batchMaxDocs = 64
)

// pubsubBatcher accumulates document updates per collection topic and publishes them
// in a single batched PushLogRequest either when batchMaxDocs is reached or after
// batchFlushInterval, whichever comes first.
type pubsubBatcher struct {
	mu      sync.Mutex
	pending map[string][]protocol.DocumentInfo // collectionID → pending docs
	timers  map[string]*time.Timer

	creator   string
	publishFn func(topic string, data []byte) error
}

func newPubsubBatcher(creator string, publishFn func(topic string, data []byte) error) *pubsubBatcher {
	return &pubsubBatcher{
		pending:   make(map[string][]protocol.DocumentInfo),
		timers:    make(map[string]*time.Timer),
		creator:   creator,
		publishFn: publishFn,
	}
}

// Add enqueues doc onto the pending batch for topic (a collectionID).
// It starts (or resets) a flush timer and flushes immediately when batchMaxDocs is reached.
func (b *pubsubBatcher) Add(topic string, doc protocol.DocumentInfo) {
	b.mu.Lock()
	b.pending[topic] = append(b.pending[topic], doc)
	full := len(b.pending[topic]) >= batchMaxDocs

	if t, ok := b.timers[topic]; ok {
		t.Stop()
		delete(b.timers, topic)
	}

	if full {
		docs := b.pending[topic]
		delete(b.pending, topic)
		b.mu.Unlock()
		b.publish(topic, docs)
		return
	}

	b.timers[topic] = time.AfterFunc(batchFlushInterval, func() {
		b.Flush(topic)
	})
	b.mu.Unlock()
}

// Flush publishes any pending documents for topic immediately and cancels the timer.
func (b *pubsubBatcher) Flush(topic string) {
	b.mu.Lock()
	docs := b.pending[topic]
	delete(b.pending, topic)
	if t, ok := b.timers[topic]; ok {
		t.Stop()
		delete(b.timers, topic)
	}
	b.mu.Unlock()

	if len(docs) > 0 {
		b.publish(topic, docs)
	}
}

// FlushAll flushes all pending topics.  Call this before shutdown.
func (b *pubsubBatcher) FlushAll() {
	b.mu.Lock()
	topics := make([]string, 0, len(b.pending))
	for topic := range b.pending {
		topics = append(topics, topic)
	}
	b.mu.Unlock()

	for _, topic := range topics {
		b.Flush(topic)
	}
}

func (b *pubsubBatcher) publish(topic string, docs []protocol.DocumentInfo) {
	req := &protocol.PushLogRequest{
		CollectionID: topic,
		Creator:      b.creator,
		Documents:    docs,
	}
	data, err := cbor.Marshal(req)
	if err != nil {
		return
	}
	_ = b.publishFn(topic, data)
}
