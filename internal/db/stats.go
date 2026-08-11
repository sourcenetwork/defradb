// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"context"
	"log/slog"
	"maps"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	ipld "github.com/ipfs/go-ipld-format"

	"github.com/sourcenetwork/corelog"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
)

// mergeStatsInterval matches the p2p reporter so the two lines of a single interval line
// up in the log.
const mergeStatsInterval = 30 * time.Second

// mergeStats are the merge-path counters reported once per interval and reset on report,
// so each line carries the rate for that interval rather than a running total.
//
// Counting is an atomic add on paths that already do storage work, so it is not a cost
// worth gating. Nothing here builds a string or allocates until the reporter runs.
type mergeStats struct {
	// creates and updates split merges by whether the document already had heads locally.
	// A merge with no local heads is creating the document.
	creates atomic.Int64
	updates atomic.Int64

	// chunkConflicts counts merge attempts abandoned for a transaction conflict, whether
	// or not a later attempt succeeded. chunkExhausted counts the chunks that ran out of
	// attempts, which is where a conflict becomes a dropped document.
	chunkConflicts atomic.Int64
	chunkExhausted atomic.Int64

	// exhaustedAtRead and exhaustedAtCommit split chunkExhausted by where the last
	// conflict was raised.
	exhaustedAtRead   atomic.Int64
	exhaustedAtCommit atomic.Int64

	// conflictKinds counts, per class of key, how many exhausted chunks had read a key of
	// that class. One exhausted chunk contributes to several entries.
	conflictKinds [datastore.ReadKindCount]atomic.Int64

	// uniqueIndexChecks counts existence checks a unique index performed before writing;
	// uniqueIndexHits counts those that found an entry already there. A hit is a
	// uniqueness violation that fails the merge outright rather than conflicting.
	uniqueIndexChecks atomic.Int64
	uniqueIndexHits   atomic.Int64

	// dropReasons counts dropped events by cause. A dropped event is a document this node
	// did not store, and the causes need different responses, so the total on its own does
	// not say what to do.
	dropMu      sync.Mutex
	dropReasons map[string]int64
}

// Causes a merge event can be dropped for. An unrecognised cause counts as dropOther, so
// that staying non-zero means this list has fallen behind the code.
const (
	dropMissingBlock   = "missingBlock"
	dropUniqueIndex    = "uniqueIndex"
	dropRetryExhausted = "retryExhausted"
	dropCollection     = "collectionNotFound"
	dropOther          = "other"
)

// mergeDropReason names why an event was dropped: the sender could not supply the DAG,
// two documents claim one indexed value, or the write kept losing to a concurrent one.
func mergeDropReason(err error) string {
	switch {
	case errors.Is(err, ipld.ErrNotFound{}):
		return dropMissingBlock
	case errors.Is(err, errors.New(errCanNotIndexNonUniqueFields)):
		return dropUniqueIndex
	case errors.Is(err, client.NewErrMaxTxnRetries(nil)):
		return dropRetryExhausted
	default:
		return dropOther
	}
}

// markDropped records an event that did not merge, under the given cause.
func (s *mergeStats) markDropped(reason string) {
	if s == nil {
		return
	}
	s.dropMu.Lock()
	defer s.dropMu.Unlock()
	if s.dropReasons == nil {
		s.dropReasons = make(map[string]int64)
	}
	s.dropReasons[reason]++
}

// drainDropReasons returns the causes counted since the last call and resets them.
func (s *mergeStats) drainDropReasons() []slog.Attr {
	s.dropMu.Lock()
	defer s.dropMu.Unlock()
	if len(s.dropReasons) == 0 {
		return nil
	}
	reasons := slices.Sorted(maps.Keys(s.dropReasons))
	attrs := make([]slog.Attr, 0, len(reasons))
	for _, reason := range reasons {
		attrs = append(attrs, corelog.Int64(reason, s.dropReasons[reason]))
	}
	clear(s.dropReasons)
	return attrs
}

// statsFor returns the counters of the database col belongs to, or nil for a collection
// this package did not construct. Every method on mergeStats tolerates a nil receiver, so
// a nil result simply means that collection is not counted.
func statsFor(col client.Collection) *mergeStats {
	if c, ok := col.(*collection); ok {
		return c.db.stats
	}
	return nil
}

// markCreateOrUpdate records whether a merge is creating the document or updating one that
// already exists locally.
func (s *mergeStats) markCreateOrUpdate(isCreate bool) {
	if s == nil {
		return
	}
	if isCreate {
		s.creates.Add(1)
		return
	}
	s.updates.Add(1)
}

// markExhausted records a chunk that ran out of attempts, along with the classes of key
// its last transaction had read. Only called on the failure path, so its cost never lands
// on a merge that commits.
func (s *mergeStats) markExhausted(atCommit bool, kinds datastore.ReadKind) {
	if s == nil {
		return
	}
	s.chunkExhausted.Add(1)
	if atCommit {
		s.exhaustedAtCommit.Add(1)
	} else {
		s.exhaustedAtRead.Add(1)
	}
	for i := range s.conflictKinds {
		if kinds&(1<<uint(i)) != 0 {
			s.conflictKinds[i].Add(1)
		}
	}
}

// markUniqueIndexCheck records a unique index existence check and whether it found an
// entry already present.
func (s *mergeStats) markUniqueIndexCheck(hit bool) {
	if s == nil {
		return
	}
	s.uniqueIndexChecks.Add(1)
	if hit {
		s.uniqueIndexHits.Add(1)
	}
}

// reportMergeStats logs the merge counters once per interval until the database context is
// cancelled. Rates are reported per interval rather than per event, which keeps the merge
// path quiet under load where per-event logging would dominate the output.
func (db *DB) reportMergeStats(ctx context.Context) {
	ticker := time.NewTicker(mergeStatsInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			db.stats.report()
		}
	}
}

func (s *mergeStats) report() {
	fields := []slog.Attr{
		corelog.Int64("creates", s.creates.Swap(0)),
		corelog.Int64("updates", s.updates.Swap(0)),
		corelog.Int64("conflicts", s.chunkConflicts.Swap(0)),
		corelog.Int64("exhausted", s.chunkExhausted.Swap(0)),
		corelog.Int64("exhaustedAtRead", s.exhaustedAtRead.Swap(0)),
		corelog.Int64("exhaustedAtCommit", s.exhaustedAtCommit.Swap(0)),
		corelog.Int64("uniqueIndexChecks", s.uniqueIndexChecks.Swap(0)),
		corelog.Int64("uniqueIndexHits", s.uniqueIndexHits.Swap(0)),
	}
	for i := range s.conflictKinds {
		fields = append(fields, corelog.Int64("read_"+datastore.ReadKindName(i), s.conflictKinds[i].Swap(0)))
	}
	log.Info("merge stats", fields...)

	// Its own line, so an interval with no drops stays quiet and the line lists only the
	// causes that occurred.
	if drops := s.drainDropReasons(); len(drops) > 0 {
		log.Error("merge drops", drops...)
	}
}
