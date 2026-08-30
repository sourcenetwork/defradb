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
	"testing"

	"github.com/stretchr/testify/require"
)

// reasonMap renders drained reasons as a map, which is the shape an assertion reads.
func reasonMap(counts []reasonCount) map[string]int64 {
	out := map[string]int64{}
	for _, rc := range counts {
		out[rc.reason] = rc.count
	}
	return out
}

// A drop is a document that was meant to be stored and was not; a skip is one deliberately
// not merged. Reporting them on one line would read policy decisions as data loss.
func TestDropAndSkipAreCountedApart(t *testing.T) {
	p := &P2P{}

	p.dropDoc("importCAR")
	p.dropDoc("importCAR")
	p.dropDoc("syncDAG")
	p.skipDoc("alreadyMerged")
	p.skipDoc("filtered")

	require.Equal(t, int64(3), p.statDroppedDocs.Load(), "drops")
	require.Equal(t, int64(2), p.statSkippedDocs.Load(), "skips")

	require.Equal(t, map[string]int64{"importCAR": 2, "syncDAG": 1}, reasonMap(p.docDropReason.drain()))
	require.Equal(t, map[string]int64{"alreadyMerged": 1, "filtered": 1}, reasonMap(p.docSkipReason.drain()))
}

// The reason set is fixed so a peer cannot grow the map, and drain resets it so each
// reported interval is a rate rather than a running total.
func TestDropReasonsDrainResets(t *testing.T) {
	p := &P2P{}
	p.dropDoc("blockDecode")
	require.Len(t, p.docDropReason.drain(), 1)
	require.Empty(t, p.docDropReason.drain(), "a drained reason set starts the next interval empty")
}
