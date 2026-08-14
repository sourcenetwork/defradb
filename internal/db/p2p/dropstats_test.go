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

// A drop is a document that was meant to be stored and was not; a skip is one deliberately
// not merged. Counting them together would report policy decisions as data loss.
func TestDropAndSkipAreCountedApart(t *testing.T) {
	p := &P2P{}

	p.dropDoc("importCAR")
	p.dropDoc("importCAR")
	p.dropDoc("syncDAG")
	p.skipDoc("alreadyMerged")

	require.Equal(t, int64(3), p.statDroppedDocs.Load(), "drops")
	require.Equal(t, int64(1), p.statSkippedDocs.Load(), "skips")

	counts := map[string]int64{}
	for _, rc := range p.docDropReason.drain() {
		counts[rc.reason] = rc.count
	}
	require.Equal(t, int64(2), counts["importCAR"])
	require.Equal(t, int64(1), counts["syncDAG"])
	require.Equal(t, int64(1), counts["skip:alreadyMerged"], "skips carry a distinct prefix")
}

// The reason set is fixed so a peer cannot grow the map, and drain resets it so each
// reported interval is a rate rather than a running total.
func TestDropReasonsDrainResets(t *testing.T) {
	p := &P2P{}
	p.dropDoc("blockDecode")
	require.Len(t, p.docDropReason.drain(), 1)
	require.Empty(t, p.docDropReason.drain(), "a drained reason set starts the next interval empty")
}
