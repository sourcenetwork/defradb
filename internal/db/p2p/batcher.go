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
	// batchWorkers is the number of goroutines that serialise and publish outbound batches.
	// Each worker handles one topic batch at a time; having 32 allows concurrent publishes
	// across collections when many attestation batches are in flight simultaneously.
	batchWorkers = 32
	// maxPubsubMessageSize is the maximum estimated wire size of a single pubsub message.
	// libp2p's default limit is 1 MB; batches that exceed this are split recursively.
	// Attestation batches carry block + transactions + logs + ales + blockSignature,
	// so individual messages can be large and this guard is load-bearing for correctness.
	maxPubsubMessageSize = 1000 * 1024
)

type publishJob struct {
	topic string
	docs  []protocol.DocumentInfo
}

// pubsubBatcher accumulates document updates per collection topic and publishes them
// in a single batched PushLogRequest either when batchMaxDocs is reached or after
// batchFlushInterval, whichever comes first. A pool of batchWorkers goroutines
// handles serialisation and publishing so the caller is never blocked by I/O.
type pubsubBatcher struct {
	mu      sync.Mutex
	pending map[string][]protocol.DocumentInfo
	timers  map[string]*time.Timer

	publishCh chan publishJob
	wg        sync.WaitGroup

	creator   string
	publishFn func(topic string, data []byte) error
}

func newPubsubBatcher(creator string, publishFn func(topic string, data []byte) error) *pubsubBatcher {
	b := &pubsubBatcher{
		pending:   make(map[string][]protocol.DocumentInfo),
		timers:    make(map[string]*time.Timer),
		publishCh: make(chan publishJob, batchWorkers*100),
		creator:   creator,
		publishFn: publishFn,
	}
	for i := 0; i < batchWorkers; i++ {
		b.wg.Add(1)
		go b.publishWorker()
	}
	return b
}

func (b *pubsubBatcher) publishWorker() {
	defer b.wg.Done()
	for job := range b.publishCh {
		b.publish(job.topic, job.docs)
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
		b.publishCh <- publishJob{topic: topic, docs: docs}
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
		b.publishCh <- publishJob{topic: topic, docs: docs}
	}
}

// Close flushes all pending topics, then shuts down the worker pool and waits for
// all in-flight publishes to complete.
func (b *pubsubBatcher) Close() {
	b.mu.Lock()
	// Collect remaining docs and stop all timers atomically. Any AfterFunc callback
	// already running will block on the lock then find pending empty and return safely.
	jobs := make([]publishJob, 0, len(b.pending))
	for topic, docs := range b.pending {
		if len(docs) > 0 {
			jobs = append(jobs, publishJob{topic: topic, docs: docs})
		}
		delete(b.pending, topic)
	}
	for topic, t := range b.timers {
		t.Stop()
		delete(b.timers, topic)
	}
	b.mu.Unlock()

	for _, job := range jobs {
		b.publishCh <- job
	}
	close(b.publishCh)
	b.wg.Wait()
}

// publish serialises docs into a PushLogRequest and calls publishFn. If the
// estimated wire size exceeds maxPubsubMessageSize the batch is split in half
// recursively until each piece fits.
func (b *pubsubBatcher) publish(topic string, docs []protocol.DocumentInfo) {
	size := 0
	for _, d := range docs {
		size += len(d.DocID) + len(d.CID) + len(d.Block) + len(d.CAR) + 50
	}
	if size > maxPubsubMessageSize && len(docs) > 1 {
		mid := len(docs) / 2
		b.publish(topic, docs[:mid])
		b.publish(topic, docs[mid:])
		return
	}

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
