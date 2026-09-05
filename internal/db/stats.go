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
)

// mergeStatsInterval matches the p2p reporter so the two lines of a single interval line
// up in the log.
const mergeStatsInterval = 30 * time.Second

// mergeStats are the merge-path counters reported once per interval and reset on report,
// so each line carries the rate for that interval rather than a running total.
//
// Counting is an atomic add on paths that already do storage work, so it is not a cost worth
// gating. The drop reasons are constants, so counting one costs a map insert and no string.
type mergeStats struct {
	// creates and updates split merges by whether the document already had heads locally.
	// A merge with no local heads is creating the document.
	creates atomic.Int64
	updates atomic.Int64

	// txnConflicts counts merge transactions abandoned for a conflict, whether or not a
	// later attempt succeeded.
	//
	// chunkExhausted counts chunks of more than one event that used their whole retry budget.
	// Such a chunk is re-run one event at a time; an event that then fails is counted under
	// dropRetryExhausted, as is a chunk of one that exhausts.
	txnConflicts   atomic.Int64
	chunkExhausted atomic.Int64

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
	dropContext        = "contextDone"
	dropOther          = "other"
)

// mergeDropReason names why an event was dropped: the sender could not supply the DAG,
// two documents claim one indexed value, or the write kept losing to a concurrent one.
//
// Match against the package sentinels. Constructing a defraError captures a stack trace,
// and this runs once per dropped event.
func mergeDropReason(err error) string {
	switch {
	case errors.Is(err, ipld.ErrNotFound{}):
		return dropMissingBlock
	case errors.Is(err, ErrCanNotIndexNonUniqueFields):
		return dropUniqueIndex
	case errors.Is(err, client.ErrMaxTxnRetries):
		return dropRetryExhausted
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return dropContext
	default:
		return dropOther
	}
}

// collectionDropReason names why an event was dropped before its collection was resolved.
// The lookup opens a transaction and reads the collection store, so not every failure here
// is a missing collection.
func collectionDropReason(err error) string {
	if errors.Is(err, client.ErrCollectionNotFound) {
		return dropCollection
	}
	return mergeDropReason(err)
}

// newMergeStats returns stats ready to record. The zero value is not usable.
func newMergeStats() *mergeStats {
	return &mergeStats{dropReasons: make(map[string]int64)}
}

// markDropped records an event that did not merge, under the given cause.
func (s *mergeStats) markDropped(reason string) {
	s.dropMu.Lock()
	defer s.dropMu.Unlock()
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

// markCreateOrUpdate records whether a merge is creating the document or updating one that
// already exists locally.
func (s *mergeStats) markCreateOrUpdate(created bool) {
	if created {
		s.creates.Add(1)
		return
	}
	s.updates.Add(1)
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
	creates := s.creates.Swap(0)
	updates := s.updates.Swap(0)
	conflicts := s.txnConflicts.Swap(0)
	exhausted := s.chunkExhausted.Swap(0)

	// Nothing to report on an idle database, or one doing only local writes.
	if creates != 0 || updates != 0 || conflicts != 0 || exhausted != 0 {
		log.Info("merge stats",
			corelog.Int64("creates", creates),
			corelog.Int64("updates", updates),
			corelog.Int64("txnConflicts", conflicts),
			corelog.Int64("chunkExhausted", exhausted),
		)
	}

	// Its own line, so an interval with no drops stays quiet and the line lists only the
	// causes that occurred.
	if drops := s.drainDropReasons(); len(drops) > 0 {
		log.Error("merge drops", drops...)
	}
}
