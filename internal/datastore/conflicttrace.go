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
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corelog"
)

// Transaction conflicts are reported by the store without naming the contended key, so
// the key has to be identified indirectly. Conflicts are detected on the read set: a
// transaction fails if a key it read was written by a transaction that committed first.
// Reads and writes are therefore counted separately, and a key that is both widely read
// and widely written is the contended one.
//
// Off unless DEFRA_TRACE_CONFLICT_KEYS is set. It records every read and write, so it is
// a diagnostic aid rather than something to leave enabled.

var log = corelog.NewLogger("datastore")

const traceReportInterval = 30 * time.Second

// traceKeyLimit bounds the tracked key count so a run cannot grow without limit between
// reports. Keys past the limit are counted but not tracked individually.
const traceKeyLimit = 200_000

var conflictTracer = newKeyTracer(os.Getenv("DEFRA_TRACE_CONFLICT_KEYS") != "")

type keyStat struct {
	txns   map[uint64]struct{}
	writes int
}

type keyTracer struct {
	enabled bool

	mu      sync.Mutex
	writes  map[string]*keyStat
	reads   map[string]*keyStat
	dropped int
	started bool
}

func newKeyTracer(enabled bool) *keyTracer {
	return &keyTracer{
		enabled: enabled,
		writes:  make(map[string]*keyStat),
		reads:   make(map[string]*keyStat),
	}
}

// record notes that txnID touched key. The first call starts the reporter, so a build
// with tracing off never spawns it.
func (t *keyTracer) record(txnID uint64, key []byte, write bool) {
	if !t.enabled {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.started {
		t.started = true
		go t.report()
	}

	set := t.reads
	if write {
		set = t.writes
	}
	stat, ok := set[string(key)]
	if !ok {
		if len(set) >= traceKeyLimit {
			t.dropped++
			return
		}
		stat = &keyStat{txns: make(map[uint64]struct{})}
		set[string(key)] = stat
	}
	stat.writes++
	stat.txns[txnID] = struct{}{}
}

type traceEntry struct {
	key        string
	readers    int
	writers    int
	readCount  int
	writeCount int
}

// snapshot returns the window's keys and clears it. A key conflicts only if it is read
// by one transaction and written by another, so entries are ordered by the smaller of
// the two counts: a key that is only read, or only written, cannot cause one.
func (t *keyTracer) snapshot() ([]traceEntry, int) {
	t.mu.Lock()
	entries := make([]traceEntry, 0, len(t.reads))
	for k, r := range t.reads {
		e := traceEntry{key: k, readers: len(r.txns), readCount: r.writes}
		if w, ok := t.writes[k]; ok {
			e.writers, e.writeCount = len(w.txns), w.writes
		}
		entries = append(entries, e)
	}
	dropped := t.dropped
	t.reads = make(map[string]*keyStat)
	t.writes = make(map[string]*keyStat)
	t.dropped = 0
	t.mu.Unlock()

	sort.Slice(entries, func(i, j int) bool {
		return min(entries[i].readers, entries[i].writers) > min(entries[j].readers, entries[j].writers)
	})
	return entries, dropped
}

// report logs the most contended keys, so each line describes one interval rather than
// everything since startup.
func (t *keyTracer) report() {
	for range time.Tick(traceReportInterval) {
		entries, dropped := t.snapshot()
		log.Info("conflict trace window",
			corelog.Int("readKeys", len(entries)),
			corelog.Int("untracked", dropped))
		for i, e := range entries {
			if i >= 10 || e.readers < 1 || e.writers < 1 {
				break
			}
			log.Info("contended key",
				corelog.Int("rank", i+1),
				corelog.Int("readerTxns", e.readers),
				corelog.Int("writerTxns", e.writers),
				corelog.Int("reads", e.readCount),
				corelog.Int("writes", e.writeCount),
				corelog.String("store", storeName(e.key)),
				corelog.String("key", describeKey(e.key)))
		}
	}
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

// traceWrites wraps store so its writes are recorded, or returns it unchanged when
// tracing is off.
func traceWrites(store corekv.ReaderWriter, txnID uint64) corekv.ReaderWriter {
	if !conflictTracer.enabled {
		return store
	}
	return &tracedStore{ReaderWriter: store, txnID: txnID}
}
