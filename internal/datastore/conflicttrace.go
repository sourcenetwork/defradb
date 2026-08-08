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
// the key has to be identified indirectly: a conflict requires one transaction to read a
// key another wrote, and writes here are read-modify-write, so a key written by many
// concurrent transactions is the shared state to look at.
//
// Off unless DEFRA_TRACE_CONFLICT_KEYS is set. It records every write, so it is a
// diagnostic aid rather than something to leave enabled.

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
	keys    map[string]*keyStat
	dropped int
	started bool
}

func newKeyTracer(enabled bool) *keyTracer {
	return &keyTracer{enabled: enabled, keys: make(map[string]*keyStat)}
}

// record notes that txnID wrote key. The first call starts the reporter, so a build with
// tracing off never spawns it.
func (t *keyTracer) record(txnID uint64, key []byte) {
	if !t.enabled {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.started {
		t.started = true
		go t.report()
	}

	stat, ok := t.keys[string(key)]
	if !ok {
		if len(t.keys) >= traceKeyLimit {
			t.dropped++
			return
		}
		stat = &keyStat{txns: make(map[uint64]struct{})}
		t.keys[string(key)] = stat
	}
	stat.writes++
	stat.txns[txnID] = struct{}{}
}

type traceEntry struct {
	key    string
	txns   int
	writes int
}

// snapshot returns the window's keys ordered by how many distinct transactions wrote
// each, and clears the window.
func (t *keyTracer) snapshot() ([]traceEntry, int) {
	t.mu.Lock()
	entries := make([]traceEntry, 0, len(t.keys))
	for k, s := range t.keys {
		entries = append(entries, traceEntry{key: k, txns: len(s.txns), writes: s.writes})
	}
	dropped := t.dropped
	t.keys = make(map[string]*keyStat)
	t.dropped = 0
	t.mu.Unlock()

	sort.Slice(entries, func(i, j int) bool { return entries[i].txns > entries[j].txns })
	return entries, dropped
}

// report logs the keys written by the most distinct transactions, so each line describes
// contention over one interval rather than since startup.
func (t *keyTracer) report() {
	for range time.Tick(traceReportInterval) {
		entries, dropped := t.snapshot()
		log.Info("conflict trace window",
			corelog.Int("distinctKeys", len(entries)),
			corelog.Int("untracked", dropped))
		for i, e := range entries {
			if i >= 10 || e.txns < 2 {
				break
			}
			log.Info("contended key",
				corelog.Int("rank", i+1),
				corelog.Int("distinctTxns", e.txns),
				corelog.Int("writes", e.writes),
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

// tracedStore records the keys a transaction writes. Reads are not recorded: a conflict
// needs a read of a key another transaction wrote, and these writes are all
// read-modify-write, so the write set already covers the shared keys.
type tracedStore struct {
	corekv.ReaderWriter
	txnID uint64
}

func (s *tracedStore) Set(ctx context.Context, key, value []byte) error {
	conflictTracer.record(s.txnID, key)
	return s.ReaderWriter.Set(ctx, key, value)
}

func (s *tracedStore) Delete(ctx context.Context, key []byte) error {
	conflictTracer.record(s.txnID, key)
	return s.ReaderWriter.Delete(ctx, key)
}

// traceWrites wraps store so its writes are recorded, or returns it unchanged when
// tracing is off.
func traceWrites(store corekv.ReaderWriter, txnID uint64) corekv.ReaderWriter {
	if !conflictTracer.enabled {
		return store
	}
	return &tracedStore{ReaderWriter: store, txnID: txnID}
}
