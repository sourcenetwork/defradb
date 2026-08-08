// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package datastore

import (
	"context"
	"encoding/hex"
	"os"
	"strconv"
	"sync"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corelog"
)

// The store reports a conflict without naming the contended key. This records what each
// open transaction reads and writes, and on a rejected commit reports the keys that
// transaction read which another wrote and committed after it started.
//
// Counting how many transactions touch a key does not work as a substitute: retries and
// lock-serialized access both inflate that count without the transactions overlapping.
//
// Off unless DEFRA_TRACE_CONFLICT_KEYS is set.

var log = corelog.NewLogger("datastore")

const (
	// traceMaxKeysPerTxn bounds what is held for one transaction. A merge reads far more
	// keys than are needed to identify a conflict, so the cap keeps a large transaction
	// from dominating memory.
	traceMaxKeysPerTxn = 20_000

	// traceHistory is how many committed transactions are kept to compare against. A
	// conflicting transaction is normally beaten by a recent commit, so a short history
	// is enough and keeps the comparison cheap.
	traceHistory = 256

	// traceMaxReported caps the keys named per conflict, since one is usually enough to
	// identify the contended structure.
	traceMaxReported = 5
)

var conflictTracer = newTracer(os.Getenv("DEFRA_TRACE_CONFLICT_KEYS") != "")

// openTxn is the state held for a transaction between its creation and its commit or
// discard.
type openTxn struct {
	reads     map[string]struct{}
	writes    map[string]struct{}
	startedAt uint64
	truncated bool
}

// commitRecord is what one committed transaction wrote, kept so a later conflict can be
// attributed to it.
type commitRecord struct {
	seq    uint64
	writes map[string]struct{}
}

type tracer struct {
	enabled bool

	mu      sync.Mutex
	open    map[uint64]*openTxn
	history []commitRecord
	seq     uint64
}

func newTracer(enabled bool) *tracer {
	return &tracer{enabled: enabled, open: make(map[uint64]*openTxn)}
}

// begin starts tracking a transaction. Read-only transactions are skipped because they
// cannot be rejected for conflict. Recording where a transaction started means a later
// conflict is only compared against commits that could have beaten it.
func (t *tracer) begin(txnID uint64, readonly bool) {
	if !t.enabled || readonly {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.open[txnID] = &openTxn{
		reads:     make(map[string]struct{}),
		writes:    make(map[string]struct{}),
		startedAt: t.seq,
	}
}

func (t *tracer) record(txnID uint64, key []byte, write bool) {
	if !t.enabled {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	txn, ok := t.open[txnID]
	if !ok {
		return
	}
	set := txn.reads
	if write {
		set = txn.writes
	}
	if len(set) >= traceMaxKeysPerTxn {
		txn.truncated = true
		return
	}
	set[string(key)] = struct{}{}
}

// commit records the outcome. On success the transaction's writes join the history so a
// later conflict can be attributed to them; on conflict its reads are compared against
// everything committed since it started.
func (t *tracer) commit(txnID uint64, conflicted bool) {
	if !t.enabled {
		return
	}
	t.mu.Lock()
	txn, ok := t.open[txnID]
	if !ok {
		t.mu.Unlock()
		return
	}
	delete(t.open, txnID)

	if !conflicted {
		t.seq++
		t.history = append(t.history, commitRecord{seq: t.seq, writes: txn.writes})
		if len(t.history) > traceHistory {
			t.history = t.history[len(t.history)-traceHistory:]
		}
		t.mu.Unlock()
		return
	}

	blamed, checked := t.blameLocked(txn)
	truncated := txn.truncated
	reads := len(txn.reads)
	t.mu.Unlock()

	if len(blamed) == 0 {
		// Either the commit that beat this transaction has aged out of the history, or
		// the contended key was reached in a way this does not record.
		log.Info("commit conflict, contended key not identified",
			corelog.Uint64("txn", txnID),
			corelog.Int("reads", reads),
			corelog.Int("commitsChecked", checked),
			corelog.Bool("readsTruncated", truncated))
		return
	}
	for _, key := range blamed {
		log.Info("commit conflict",
			corelog.Uint64("txn", txnID),
			corelog.Int("reads", reads),
			corelog.String("store", storeName(key)),
			corelog.String("contendedKey", describeKey(key)))
	}
}

// blameLocked returns the keys txn read that a later commit wrote, and how many commits
// were considered. The caller must hold the lock.
func (t *tracer) blameLocked(txn *openTxn) ([]string, int) {
	blamed := make([]string, 0, traceMaxReported)
	checked := 0
	for _, c := range t.history {
		if c.seq <= txn.startedAt {
			continue
		}
		checked++
		for key := range c.writes {
			if _, read := txn.reads[key]; read {
				blamed = append(blamed, key)
				if len(blamed) >= traceMaxReported {
					return blamed, checked
				}
			}
		}
	}
	return blamed, checked
}

// discard drops a transaction's state. Without it an abandoned transaction would be held
// until the process exits.
func (t *tracer) discard(txnID uint64) {
	if !t.enabled {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.open, txnID)
}

// storeName maps the leading namespace byte to the store it belongs to.
func storeName(key string) string {
	if key == "" {
		return "empty"
	}
	switch key[0] {
	case systemStoreKey:
		return "system"
	case dataStoreKey:
		return "data"
	case headStoreKey:
		return "head"
	case peerStoreKey:
		return "peer"
	case encStoreKey:
		return "enc"
	case blockStoreKey:
		if len(key) > 1 && key[1] == toMergeIndexPrefix {
			return "block/to-merge"
		}
		return "block"
	default:
		return "other:" + strconv.Quote(key[:1])
	}
}

// describeKey renders a key for a log line, keeping it printable and short. Keys carry
// binary segments, so anything unprintable is hex encoded.
func describeKey(key string) string {
	const maxLen = 72
	printable := true
	for i := 0; i < len(key); i++ {
		if key[i] < 0x20 || key[i] > 0x7e {
			printable = false
			break
		}
	}
	out := key
	if !printable {
		out = "0x" + hex.EncodeToString([]byte(key))
	}
	if len(out) > maxLen {
		out = out[:maxLen] + "..."
	}
	return out
}

// tracedStore records the keys a transaction reads and writes, including keys reached by
// scanning: the store adds those to the read set too, so leaving them out would hide a
// whole class of conflict.
type tracedStore struct {
	corekv.ReaderWriter
	txnID uint64
}

func (s *tracedStore) Set(ctx context.Context, key, value []byte) error {
	conflictTracer.record(s.txnID, key, true)
	return s.ReaderWriter.Set(ctx, key, value)
}

func (s *tracedStore) Delete(ctx context.Context, key []byte) error {
	conflictTracer.record(s.txnID, key, true)
	return s.ReaderWriter.Delete(ctx, key)
}

func (s *tracedStore) Get(ctx context.Context, key []byte) ([]byte, error) {
	conflictTracer.record(s.txnID, key, false)
	return s.ReaderWriter.Get(ctx, key)
}

func (s *tracedStore) Has(ctx context.Context, key []byte) (bool, error) {
	conflictTracer.record(s.txnID, key, false)
	return s.ReaderWriter.Has(ctx, key)
}

func (s *tracedStore) Iterator(ctx context.Context, opts corekv.IterOptions) (corekv.Iterator, error) {
	iter, err := s.ReaderWriter.Iterator(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &tracedIterator{Iterator: iter, txnID: s.txnID}, nil
}

// tracedIterator records each key the iterator lands on. Next and Seek are the only ways
// to reach a new position, so recording there covers every key visited.
type tracedIterator struct {
	corekv.Iterator
	txnID uint64
}

func (i *tracedIterator) Next() (bool, error) {
	ok, err := i.Iterator.Next()
	if ok {
		conflictTracer.record(i.txnID, i.Iterator.Key(), false)
	}
	return ok, err
}

func (i *tracedIterator) Seek(key []byte) (bool, error) {
	ok, err := i.Iterator.Seek(key)
	if ok {
		conflictTracer.record(i.txnID, i.Iterator.Key(), false)
	}
	return ok, err
}

// traceKeys wraps store so the transaction's keys are recorded, or returns it unchanged
// when there is nothing to record.
func traceKeys(store corekv.ReaderWriter, txnID uint64, readonly bool) corekv.ReaderWriter {
	if !conflictTracer.enabled || readonly {
		return store
	}
	return &tracedStore{ReaderWriter: store, txnID: txnID}
}
